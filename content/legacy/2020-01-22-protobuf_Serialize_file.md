---
title: protobuf 序列化到文件及反序列化
date: '2020-01-22T20:29:39+08:00'
url: /2020/01/22/protobuf_Serialize_file/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 项目应用 {#项目应用}

游戏中，将对局的数据保留下来，用于对局回顾及debug等用途，由于协议采用PB，故以二进制的pb格式写入文件，在使用该对局内容的时候，按照格式反序列化出来用于播放对局、压测数据构造等。

涉及的部分： pvp服务器，产生对局数据，然后通过路由发送到recordsvr，一个专门写文件的服务器，写完文件后，使用时对文件解析。

文件的格式: head-data-head-data….

相关的协议如下：  

```plain
message RecordReq
{
    optional uint64 game_id = 1;
    optional string record_name = 2;
    optional bool finish = 3;
    optional RecordPkg Pkg = 4;
}

message RecordPkg
{
    optional uint32 time_offfect = 1;
    optional uint32 cmd = 2;
    optional bytes fragment_data = 3;
}

message RecordHead
{
    required fixed32 cmd = 1;  
    required fixed32 len = 2;
    required fixed32 time_offect = 3;
}
```

### pvp对局服务器，产生数据 {#pvp对局服务器，产生数据}

```plain
void CPvPGame::SendRecord(uint16_t uiCmd, const Message* msg, bool bFinish/* = false*/)
{
    unsigned int uiNow = CSingletonPvPDateTime::instance()->GetSec();
    uint32_t uiTimeOffset = uiNow - m_battleInfo.battle_base.battle_start_time;

    RecordReq msgReq;
    RecordPkg* pPkg = msgReq.mutable_pkg();
    if (NULL == pPkg)
    {
        return;
    }
    msgReq.set_game_id(GetGameId());
    msgReq.set_record_name(GetRecordName());
    msgReq.set_finish(bFinish);
    pPkg->set_time_offfect(uiTimeOffset);
    pPkg->set_cmd(uiCmd);


    if (GAME_RECORD_START == uiCmd || GAME_RECORD_END == uiCmd)
    {
        GameDataInfo gameData;
        FillGameInfo(&gameData);
        //修正starttime
        gameData.mutable_game_init_info()->set_start_time(m_uiStartTime);
        gameData.SerializeToString(pPkg->mutable_fragment_data());
    }
    else if (NULL != msg)
    {
        msg->SerializeToString(pPkg->mutable_fragment_data());
    }
    else
    {
        return;
    }

    int iRet = 0;
    iRet = CSingletonPvPProcessor::instance()->
        SendMsgToAllServerByIdc(GetOwer(), RecordSvrCmd::RECORD_REQ, msgReq, SERVER_FAMILY_RECORDSVR);
```

### 录像服务器，专门写文件 {#录像服务器，专门写文件}

```plain
std::stringstream sstrNameTmp;
sstrNameTmp << szRecordDir << "/tmp_" << msgReq.game_id() << "_tmp";

RecordHead rHead;
rHead.set_cmd(msgReq.pkg().cmd());
rHead.set_len(msgReq.pkg().fragment_data().size());
rHead.set_time_offect(msgReq.pkg().time_offfect());

//std::fstream output(sstrNameTmp.str().c_str(), std::ios::out | std::ios::binary | std::ios::app);
std::fstream& output = OpenFile(sstrNameTmp.str().c_str());

std::string strout;
rHead.SerializeToString(&strout);
output << strout;
if (msgReq.pkg().has_fragment_data())
{
    output << msgReq.pkg().fragment_data();
}

//output.close();

LOG_DEBUG(0, 0, "write record|%lu|%s", msgReq.game_id(), sstrNameTmp.str().c_str());
LOG_DEBUG(0, 0, "write record|%lu|%d|%s", msgReq.game_id(), rHead.ByteSize(), rHead.ShortDebugString().c_str());

if (msgReq.finish())
{
    CloseFile(sstrNameTmp.str().c_str());
    rename(sstrNameTmp.str().c_str(), strLogName.c_str());
    LOG_DEBUG(0, 0, "rename|%lu|%s|%s", msgReq.game_id(), sstrNameTmp.str().c_str(), strLogName.c_str());
}
```

### 解析对局数据，用于回放或性能测试 {#解析对局数据，用于回放或性能测试}

```plain
  bool Robot::ReadData()
  {
m_IsInitGameDataInfo = true;
      char* writebuff=RunTime::GetInst()->getwritebuff();
      long filelength=RunTime::GetInst()->getfilelength();
      DEBUG("enter ReadData");
      DEBUG("writebuff len:%d,length:%ld",(int)strlen(writebuff),filelength);

/*
message RecordHead
{
	required fixed32 cmd = 1;  //cmd 为0的时候，包体为GameDataInfo结构
	required fixed32 len = 2;
	required fixed32 time_offect = 3;
}
*/
// 先 算一个包头
RecordHead recHead;
recHead.set_cmd(869);			// GAME_RECORD_START
recHead.set_len(1024);			//  fixed32  any 32 bit
recHead.set_time_offect(1024);	// fixed32; any 32 bit
std::string strout;
recHead.SerializeToString(&strout);
int HeadSize = strout.size();

recHead.ParseFromArray(writebuff, HeadSize );
DEBUG("RecordHead: %s", recHead.ShortDebugString().c_str());

m_gameDataInfo.ParseFromArray(writebuff + HeadSize, recHead.len());
DEBUG("m_gameDataInfo: %s", m_gameDataInfo.ShortDebugString().c_str());
```
