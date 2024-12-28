import yaml
import sys

class Conf():
    def __init__(self):
        self.conf = {}
        with open("conf/mcore_config.yaml") as f:
            self.conf = yaml.load(f, Loader=yaml.FullLoader)

        self.users =  self.conf.get("users")
        if len(self.users) < 1:
            print("need users in config")
            sys.exit(1) 

    @property
    def task_type(self) -> int:
        return self.conf.get("task_type", 2)

    @property
    def mcore_addr(self) -> str:
        return self.conf.get("mcore_addr", "http://localhost:8080")

    @property
    def ytld_mcore_secret(self) -> str:
        return self.conf.get("ytld_mcore_secret", "test")

    @property
    def path2content( self ) ->str: 
        return self.conf.get("path2content", "temp")
    
    @property
    def url2content( self ) ->str: 
        return self.conf.get("url2content", "")
    
    @property
    def retries( self ) ->int: 
        return self.conf.get("retries", 20)
    
    def get_user_conf( self, user: str) -> dict:
        for u in self.users:
            if u["name"] == user:
                return u

        return None
            
    