import yaml
import sys

class Conf():
    def __init__(self):
        self.conf = {}
        with open("conf/tr_mng_config.yaml") as f:
            self.conf = yaml.load(f, Loader=yaml.FullLoader)

    @property
    def trhost( self ) ->str:
        return self.conf.get("trhost", "transmission")

    @property
    def trport( self ) ->int:
        return self.conf.get("trport", 9091)

    @property
    def tdownloaddir( self ) ->int:
        return self.conf.get("tdownload_dir", "/downloads/complete/")


    @property
    def task_type(self) -> int:
        return self.conf.get("taskType", 5)

    @property
    def mcore_addr(self) -> str:
        return self.conf.get("mcoreAddr", "http://localhost:8080")

    @property
    def mcore_secret(self) -> str:
        return self.conf.get("mcoreSecret", "test")