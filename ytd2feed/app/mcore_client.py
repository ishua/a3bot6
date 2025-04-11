import requests
import sys

class McoreClient:
    def __init__(self, addr:str, task_type:int, secret:str):
        self.addr = addr
        self.task_type = task_type
        self.secret = secret

    def health(self) -> bool:
        try:
            r = requests.get(self.addr + "/health/")
        except Exception as e:
            print(e)
            sys.exit(1)
        if r.status_code != 200:
            print("health status code:", r.status_code)
            print("health body", r.request.body)
            return False
        return True

    def get_task(self) -> dict:
        headers =  {
            'content-type': 'application/json',
            'secret': self.secret
        }
        try:
            r = requests.post(self.addr + "/get-task/", json={'taskType': self.task_type}, headers=headers)
        except Exception as e:
            print(e)
            return {}

        if r.status_code != 200:
            print("can't connect to get-task")
            return {}

        res = r.json()
        if res['status'] == "no tasks":
            return {}
        if res['status'] != "OK":
            print("something went wrong:",res["error"])
            return {}
        if res['data'] is None:
            return {}
        return res['data']

    def report_task(self, task_id:int, status: int, text_msg: str) -> bool:
        headers =  {
            'content-type': 'application/json',
            'secret': self.secret
        }
        data = {
            "taskId": task_id,
            "status": status,
            "textMsg": text_msg
        }
        try:
            r = requests.post(self.addr + "/report-task/", json=data, headers=headers)
        except Exception as e:
            print(e)
            return False

        if r.status_code != 200:
            print("report status err")
            print(r.json())
            return False

        return  True

    def check_and_report(self, task: dict) -> bool:
        print("check task")
        if task.get("id") is None:
            print("TODO crit need send a message task")
            sys.exit(1)
        print("taskid:",str(task["id"]))
        if task.get("taskData") is None:
            errorMsg =  "ytdl taskData is empty"
            print("taskid:", str(task["id"]) ,errorMsg)
            self.report_task(task["taskData"], 2, errorMsg)
            return False
        if task["taskData"].get("ytdl") is None:
            errorMsg =  "ytdl taskData.ytdl is empty"
            print("taskid:", str(task["id"]) ,errorMsg)
            self.report_task(task["taskData"]["ytdl"], 2, errorMsg)
            return False
        if task["taskData"]["ytdl"].get("link") is None:
            errorMsg =  "ytdl taskData.ytdl.link is empty"
            print("taskid:", str(task["id"]) ,errorMsg)
            self.report_task(task["taskData"]["ytdl"]["link"], 2, errorMsg)
            return False
        if task["taskData"]["ytdl"].get("userName") is None:
            errorMsg =  "ytdl taskData.ytdl.userName is empty"
            print("taskid:", str(task["id"]) ,errorMsg)
            self.report_task(task["taskData"]["ytdl"]["userName"], 2, errorMsg)
            return False

        return self.report_task(task["id"], 3, "ytdl get the job: " + str(task["id"]))


