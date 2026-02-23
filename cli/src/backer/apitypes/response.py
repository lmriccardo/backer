import enum

from pydantic import BaseModel, Field, field_validator
from typing import Optional, List

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
    ok      : bool
    version : VersionResponse

class ResponseStatus(enum.Enum):
    Success = 0
    Failure = 1

class BaseResponse(ResponseModel):
    status  : ResponseStatus
    message : Optional[str] = None
    errors  : List[str]     = Field(default_factory=list)

    @field_validator('errors', mode="before")
    @classmethod
    def errors_none_to_list(cls, v : List[str] | None) -> List[str]:
        if v is None: return []
        if not isinstance(v, list):
            raise ValueError("errors field must be a list")
        return v

    def is_success(self) -> bool:
        return self.status == ResponseStatus.Success

    def __errors_to_str(self) -> str:
        if self.is_success(): return ""
        if len(self.errors) == 1:
            return f"[red]Error[/red]: {self.errors[0]}"

        result = "[red]Errors[/red]:\n"
        for idx, error in enumerate(self.errors):
            result += f"  • ([cyan]{idx}[/cyan]) {error}\n"
        return result

    def __msg_to_str(self) -> None:
        if not self.is_success(): return ""
        return f"[green]{self.message}[/green]"
    
    def __str__(self) -> str:
        if self.is_success(): return self.__msg_to_str()
        else: return self.__errors_to_str()
