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

        print("mcore health ok")
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
            print("can't connect to get-task, status: ", r.status_code)
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
        if task.get("id") is None:
            print("task have not id")
            sys.exit(1)
        print("check task id =", task["id"])
        error_msg = ""
        if task.get("taskData") is None:
            error_msg =  "transmission taskData is empty"
        elif task["taskData"].get("tr") is None:
            error_msg =  "transmission taskData.tr is empty"
        elif task["taskData"]["tr"].get("command") is None:
            error_msg =  "transmission taskData.tr is empty"
        elif task["taskData"]["tr"]["command"] == "add":
            if task["taskData"]["tr"].get("torrentUrl") is None:
                error_msg =  "transmission taskData.tr.torrentUrl is empty"
            elif task["taskData"]["tr"].get("folderPath") is None:
                error_msg =  "transmission taskData.tr.folderPath is empty"
        elif task["taskData"]["tr"]["command"] == "del":
            if task["taskData"]["tr"].get("torrentId") is None:
                error_msg =  "transmission taskData.tr.torrentId is empty"
        elif task["taskData"]["tr"]["command"] != "list":
            error_msg =  "transmission taskData.tr.command not add/del/list"

        if error_msg != "":
            print("taskid:", str(task["id"]) ,error_msg)
            self.report_task(task["id"], 2, error_msg)
            return False

        return self.report_task(task["id"], 3, "transmission get the job: " + str(task["id"]))

    def health_reported(self, task: dict) -> bool:
        if task.get("id") is None:
            print("TODO crit need send a message task")
            sys.exit(1)
        if task.get("taskData") is None:
            return False
        if task["taskData"].get("health") is None:
            return False
        if len(task["taskData"].get("health"))  == 0 :
            return False

        self.report_task(task["id"], 4, "tr_mng is healthy: " + str(task["id"]))
        return True