import requests

class TaskWorker:
    def __init__(self, addr:str, task_type:int, secret:str):
        self.addr = addr
        self.task_type = task_type
        self.secret = secret

    def get_task(self) -> dict:
        headers =  {
            'content-type': 'application/json',
            'secret': self.secret
        }
        r = requests.post(self.addr + "/get-task/", data={'taskType': self.task_type}, headers=headers)
        if r.status_code != 200:
            print("can't connect to get-task")
            return {}

        res = r.json()
        if res['status'] != "OK":
            print("something went wrong:",res["error"])
            return {}
        if res['data'] is None:
            return {}
        return res['data']

    def report_task(self, task_id:int, status: str, text_msg: str):
        headers =  {
            'content-type': 'application/json',
            'secret': self.secret
        }
        data = {
            "taskId": task_id,
            "status": status,
            "textMsg": text_msg
        }
        r = requests.post(self.addr + "/report-task/", data=data, headers=headers)
        if r.status_code != 200:
            print("report status err")
            print(r.json())
