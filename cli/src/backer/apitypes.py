from pydantic import BaseModel

class ResponseModel(BaseModel): ...

class VersionResponse(ResponseModel):
    api        : int
    commit     : str
    date       : str
    go_version : str
    version    : str

    def __str__(self) -> str:
        return (
            f"Daemon Version: backerd {self.version} " +
            f"({self.commit}) built {self.date} " +
            f"go={self.go_version}"
        )

class HealthzResponse(ResponseModel):
    ok: bool
    version: VersionResponse