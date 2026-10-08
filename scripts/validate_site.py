"""Check migration invariants and prevent accidental publication of retired posts."""
from pathlib import Path
from html.parser import HTMLParser
from urllib.parse import unquote, urlsplit
import json, sys, hashlib
from collections import Counter

root = Path(__file__).resolve().parents[1]
public = Path(sys.argv[1]) if len(sys.argv) > 1 else root / 'public'
report = json.loads((root / 'migration-report.json').read_text())
retired = ('tencent_20th', 'macroeconomics_2024')
errors = []

class ArticleParser(HTMLParser):
    def __init__(self):
        super().__init__(); self.images = []; self.pre = 0; self.links = []; self.code_text = []; self.in_pre = 0; self.ids = set()
    def handle_starttag(self, tag, attrs):
        attr = dict(attrs)
        if 'id' in attr: self.ids.add(attr['id'])
        if tag == 'img': self.images.append(attr.get('src', ''))
        if tag == 'pre': self.pre += 1; self.in_pre += 1; self.code_text.append('')
        if tag == 'a': self.links.append(attr.get('href', ''))
    def handle_endtag(self, tag):
        if tag == 'pre': self.in_pre -= 1
    def handle_data(self, data):
        if self.in_pre: self.code_text[-1] += data

for post in report:
    p = public / unquote(post['url']).strip('/') / 'index.html'
    if not p.exists(): errors.append('Missing original URL: ' + post['url']); continue
    parser = ArticleParser(); parser.feed(p.read_text())
    if len(parser.images) < post['images']: errors.append('Image count decreased: ' + post['url'])
    if parser.pre < post['code_blocks']: errors.append('Code blocks decreased: ' + post['url'])
    actual_hashes = Counter(hashlib.sha256('\n'.join(line if line.strip() else '' for line in text.replace('\r\n','\n').strip('\n').split('\n')).encode()).hexdigest() for text in parser.code_text)
    if Counter(post.get('code_hashes',[])) - actual_hashes: errors.append('Code content changed: ' + post['url'])
    if set(post.get('heading_ids',[])) - parser.ids: errors.append('Heading anchors changed: ' + post['url'])
    for src in parser.images:
        parsed = urlsplit(src)
        if parsed.scheme or parsed.netloc or not src or src.startswith('data:'): continue
        image = public / unquote(parsed.path).lstrip('/') if src.startswith('/') else p.parent / unquote(parsed.path)
        if not image.exists(): errors.append('Missing local image: ' + post['url'] + ' -> ' + src)

for p in public.rglob('*.html'):
    text = p.read_text()
    if any('/' + slug + '/' in text for slug in retired): errors.append('Retired article referenced: ' + str(p.relative_to(public)))

index = json.loads((public / 'index.json').read_text())
urls = [post['url'] for post in index]
if len(urls) != len(set(urls)): errors.append('Duplicate article URLs')
if not all(post['url'] in urls for post in report if not post.get('legacy')): errors.append('Search index omits original articles')
if any(post['url'] in urls for post in report if post.get('legacy')): errors.append('Unlisted legacy article appeared in the search index')
if any(any(slug in url for slug in retired) for url in urls): errors.append('Retired article in search index')
if not (public / 'archives/index.html').exists(): errors.append('Archive missing')
for url, expected in (('/2024/03/26/vedio_codec_01/', '码控'), ('/2019/01/04/vmstat/', '第一行数字')):
    p = public / url.strip('/') / 'index.html'
    if p.exists() and expected not in p.read_text(): errors.append('Recovered article unexpectedly empty: ' + url)

if errors:
    print('\n'.join(errors)); sys.exit(1)
print(f'PASS: {sum(not p.get("legacy") for p in report)} active articles, {len(report)} original URLs, images, code blocks, archive and search; retired articles excluded.')
