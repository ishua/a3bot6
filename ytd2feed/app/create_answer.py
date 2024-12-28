import json
from urllib.parse import urlparse

# type Msg struct {
# 	Command          string `json:"command"`
# 	UserName         string `json:"userName"`
# 	MsgId            int    `json:"msgId"`
# 	ReplyToMessageID int    `json:"replyToMessageID"`
# 	ChatId           int64  `json:"chatId"`
# 	Text             string `json:"text"`
# 	ReplyText        string `json:"replyText"`
# }

class Answer:
    def __init__( self, payload: str, redis: str, channel: str):
        self.msg = json.loads(payload)
        self.redis = redis
        self.channel = channel
        
    def setReply( self, reply: str):
        self.msg["ReplyText"] = reply

    def getReply( self ) -> str:
        return self.msg.get("ReplyText", "")
    
    def getUserName( self) -> str:
        return self.msg.get("userName", "")
    
    def validateError( self ) -> bool:
        if self.msg.get("text", "") == "":
            self.setReply("there is no command")
            return True

        c = self.msg["text"].split(" ")
        if len(c) < 2:
            self.setReply("there is no command link")
            return True
        
        _uri = urlparse(c[1])
        if _uri.hostname is None:
            self.setReply("no link")
            return True
        
        if not (_uri.hostname == "youtube.com" or _uri.hostname == "www.youtube.com" or _uri.hostname == "youtu.be"):
            self.setReply("there is no youtobe link")
            return True
        
        return False
    
    def getUrl( self ) -> str:
        c = self.msg["text"].split(" ")
        return c[1]


    def send( self ):
        if self.getReply == "":
            print("some error with send answer")
        r = redis.Redis(self.redis)
        r.publish(self.channel, json.dumps(self.msg))
        