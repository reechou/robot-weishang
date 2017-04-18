package controller

import (
	"fmt"
	"strings"
	
	"github.com/reechou/holmes"
	"github.com/reechou/robot-weishang/robot_proto"
	"github.com/reechou/robot-weishang/ext"
)

func (self *Logic) HandleReceiveMsg(msg *robot_proto.ReceiveMsgInfo) {
	holmes.Debug("receive robot msg: %v", msg)
	switch msg.BaseInfo.ReceiveEvent {
	case robot_proto.RECEIVE_EVENT_MSG:
		self.handleMsg(msg)
	case robot_proto.RECEIVE_EVENT_ADD, robot_proto.RECEIVE_EVENT_ADD_FRIEND:
		self.handleAddFriend(msg)
	}
}

func (self *Logic) handleMsg(msg *robot_proto.ReceiveMsgInfo) {
	switch msg.BaseInfo.FromType {
	case robot_proto.FROM_TYPE_PEOPLE:
		for _, v := range self.cfg.NotifyList {
			if msg.BaseInfo.FromNickName == v {
				self.masterMsg(msg)
				return
			}
		}
		self.transferMsg(msg)
	}
}

func (self *Logic) handleAddFriend(msg *robot_proto.ReceiveMsgInfo) {
	robotHost := self.getRobotHost(msg.BaseInfo.WechatNick)
	if robotHost == "" {
		holmes.Error("get robot host == nil")
		return
	}
	var sendReq robot_proto.SendMsgInfo
	sendReq.SendMsgs = append(sendReq.SendMsgs, robot_proto.SendBaseInfo{
		WechatNick: msg.BaseInfo.WechatNick,
		ChatType:   msg.BaseInfo.FromType,
		UserName:   msg.BaseInfo.FromUserName,
		NickName:   msg.BaseInfo.FromNickName,
		MsgType:    robot_proto.RECEIVE_MSG_TYPE_TEXT,
		Msg:        self.cfg.AddMsg,
	})
	err := self.robotExt.SendMsgs(robotHost, &sendReq)
	if err != nil {
		holmes.Error("add friend send msg[%v] error: %v", sendReq, err)
	}
	
	var msgDetail string
	msgDetail = fmt.Sprintf("%s 新添加好友[%s]", msg.BaseInfo.FromUserName, msg.BaseInfo.FromNickName)
	
	for _, v := range self.cfg.NotifyList {
		var sendReq robot_proto.SendMsgInfo
		sendReq.SendMsgs = append(sendReq.SendMsgs, robot_proto.SendBaseInfo{
			WechatNick: msg.BaseInfo.WechatNick,
			ChatType:   robot_proto.FROM_TYPE_PEOPLE,
			NickName:   v,
			MsgType:    robot_proto.RECEIVE_MSG_TYPE_TEXT,
			Msg:        msgDetail,
		})
		err := self.robotExt.SendMsgs(robotHost, &sendReq)
		if err != nil {
			holmes.Error("add friend send msg[%v] error: %v", sendReq, err)
		}
	}
}

func (self *Logic) masterMsg(msg *robot_proto.ReceiveMsgInfo) {
	robotHost := self.getRobotHost(msg.BaseInfo.WechatNick)
	if robotHost == "" {
		holmes.Error("get robot host == nil")
		return
	}
	msgInfo := strings.Fields(msg.Msg)
	if len(msgInfo) < 2 {
		holmes.Error("msg: %s error format", msg.Msg)
		return
	}
	if !strings.HasPrefix(msgInfo[0], "@") {
		holmes.Error("user id: %s error.", msgInfo[0])
		return
	}
	var sendMsg string
	if len(msgInfo) > 2 {
		sendMsg = strings.Join(msgInfo[1:], " ")
	} else {
		sendMsg = msgInfo[1]
	}
	var sendReq robot_proto.SendMsgInfo
	sendReq.SendMsgs = append(sendReq.SendMsgs, robot_proto.SendBaseInfo{
		WechatNick: msg.BaseInfo.WechatNick,
		ChatType:   robot_proto.FROM_TYPE_PEOPLE,
		UserName:   msgInfo[0],
		MsgType:    robot_proto.RECEIVE_MSG_TYPE_TEXT,
		Msg:        sendMsg,
	})
	err := self.robotExt.SendMsgs(robotHost, &sendReq)
	if err != nil {
		holmes.Error("add friend send msg[%v] error: %v", sendReq, err)
	}
}

func (self *Logic) transferMsg(msg *robot_proto.ReceiveMsgInfo) {
	if msg.BaseInfo.UserName == msg.BaseInfo.FromUserName {
		return
	}
	
	if strings.Contains(msg.Msg, "我通过了你的朋友验证请求") {
		return
	}
	
	robotHost := self.getRobotHost(msg.BaseInfo.WechatNick)
	if robotHost == "" {
		holmes.Error("get robot host == nil")
		return
	}
	var msgDetail string
	if msg.MsgType == robot_proto.RECEIVE_MSG_TYPE_TRANSFER || msg.MsgType == robot_proto.RECEIVE_MSG_TYPE_RED_PACKET {
		msgDetail = fmt.Sprintf("%s %s ---来自[%s]", msg.BaseInfo.FromUserName, msg.Msg, msg.BaseInfo.FromNickName)
	} else {
		msgDetail = fmt.Sprintf("%s %s", msg.BaseInfo.FromUserName, msg.Msg)
	}
	
	for _, v := range self.cfg.NotifyList {
		var sendReq robot_proto.SendMsgInfo
		sendReq.SendMsgs = append(sendReq.SendMsgs, robot_proto.SendBaseInfo{
			WechatNick: msg.BaseInfo.WechatNick,
			ChatType:   robot_proto.FROM_TYPE_PEOPLE,
			NickName:   v,
			MsgType:    robot_proto.RECEIVE_MSG_TYPE_TEXT,
			Msg:        msgDetail,
		})
		err := self.robotExt.SendMsgs(robotHost, &sendReq)
		if err != nil {
			holmes.Error("add friend send msg[%v] error: %v", sendReq, err)
		}
	}
}

func (self *Logic) getRobotHost(robot string) string {
	robotReq := &ext.GetRobotReq{
		RobotWx: robot,
	}
	robotInfo, err := self.robotCExt.GetRobot(robotReq)
	if err != nil {
		holmes.Error("get robot from controller error: %v", err)
		return ""
	}
	return fmt.Sprintf("%s%s", robotInfo.Ip, robotInfo.OfPort)
}
