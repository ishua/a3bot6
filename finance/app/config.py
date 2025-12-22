import yaml
import sys

class Conf():
    def __init__(self):
        self.conf = {}
        with open("conf/fin_config.yaml") as f:
            self.conf = yaml.load(f, Loader=yaml.FullLoader)

    @property
    def task_type(self) -> int:
        return self.conf.get("taskType", 6)

    @property
    def mcore_addr(self) -> str:
        return self.conf.get("mcoreAddr", "http://localhost:8080")

    @property
    def mcore_secret(self) -> str:
        return self.conf.get("mcoreSecret", "test")
