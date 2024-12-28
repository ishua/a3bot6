import unittest
from app import Answer
import json

def _getPayload(  msg: dict ) -> str:
        return json.dumps(msg)

msg = {}

class TestAnswer(unittest.TestCase):
    def test_upper(self):
        msg = {
                "command": "/y2d",
                "userName": "UserName",
                "msgId": 2,
                "replyToMessageID": 1,
                "chatId": 10,
                "text": "y https://www.youtube.com/watch?v=O11dfJVJusk"
        }

    def test_answer_init(self):
        a = Answer(_getPayload(msg))
        self.assertDictEqual(msg, a.msg, "init answer")

    def test_validateError_no_text(self):
        _msg = msg.copy()
        _msg["text"] = ""
        a = Answer(_getPayload(_msg))
        r = a.validateError()
        self.assertTrue(r)
        self.assertEqual("there is no command", a.getReply())

    def test_validateError_no_link(self):
        _msg = msg.copy()
        _msg["text"] = "y"
        a = Answer(_getPayload(_msg))
        r = a.validateError()
        self.assertTrue(r)
        self.assertEqual("there is no command link", a.getReply())

    def test_validateError_wrong_link(self):
        _msg = msg.copy()
        _msg["text"] = "y y"
        a = Answer(_getPayload(_msg))
        r = a.validateError()
        self.assertTrue(r)
        self.assertEqual("no link", a.getReply())

    def test_validateError_no_youtobe_link(self):
        _msg = msg.copy()
        _msg["text"] = "y https://google.com"
        a = Answer(_getPayload(_msg))
        r = a.validateError()
        self.assertTrue(r)
        self.assertEqual("there is no youtobe link", a.getReply())

    def test_validateError_youtobe_link(self):
        _msg = msg.copy()
        _msg["text"] = "y https://www.youtube.com/watch?v=O11dfJVJusk"
        a = Answer(_getPayload(_msg))
        r = a.validateError()
        print(a.getReply())
        self.assertFalse(r)
        self.assertEqual("", a.getReply())
        
         

if __name__ == '__main__':
    unittest.main()