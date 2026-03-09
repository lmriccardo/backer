from __future__ import annotations

import os
import enum

from typing import Optional, List, Union, Dict, Any
from pydantic import (
    BaseModel, Field, field_validator,
    model_validator, ConfigDict,
    HttpUrl, EmailStr
)

CronField = Optional[Union[int,str]]

class RequestModel(BaseModel): ...

class CaseInsensitiveEnum(str, enum.Enum):
    @classmethod
    def _missing_(cls, value):
        if not isinstance(value, str): return None
        for member in cls:
            if member.value.lower() == value.lower():
                return member
            
class DeleteType(CaseInsensitiveEnum):
    delete_after    = "after"
    delete_before   = "before"
    delete_delay    = "delay"
    delete_during   = "during"
    delete_excluded = "excluded"

class RemoteDestCfg(BaseModel):
    model_config = ConfigDict(extra="forbid")

    module: str # The remote rsync module
    folder: str # The remote folder under the modul

class RemoteCfg(BaseModel):
    model_config = ConfigDict(extra="forbid")

    host     : str
    port     : int = Field(default=873, ge=1, le=65535)
    user     : str
    password : str
    dest     : RemoteDestCfg

    @field_validator("password", mode="before")
    @classmethod
    def expandenvs( cls, path: str ) -> str:
        """ Expand the environment variable if present """
        return os.path.expandvars( path )
    
class RsyncOptions(BaseModel):
    model_config = ConfigDict(extra="forbid", validate_default=True)

    compress        : bool                 = False
    verbose         : bool                 = True
    show_progress   : bool                 = True
    itemize_changes : bool                 = False
    delete          : Optional[DeleteType] = None
    keep_specials   : bool                 = False
    keep_devices    : bool                 = False

    @model_validator(mode='after')
    def default_delete_mode(self) -> 'RsyncOptions':
        if self.delete is None:
            self.delete = DeleteType.delete_after
        return self
    
class RsyncCfg(BaseModel):
    model_config = ConfigDict(extra="forbid", validate_default=True)

    exclude_output_folder : Optional[str]          = None
    exclude_from          : Optional[str]          = None
    excludes              : List[str]              = Field(default_factory=list)
    includes              : List[str]              = Field(default_factory=list)
    sources               : List[str]              = Field(default_factory=list, min_length=1)
    options               : Optional[RsyncOptions] = None

    @field_validator(
        "exclude_output_folder",
        "exclude_from",
        "excludes",
        "includes",
        "sources",
        mode="before",
    )
    @classmethod
    def expandenvs( cls, path: str | None | List[str] ) -> str | None | List[str]:
        """ Expand the environment variable if present """
        if not path and not isinstance(path, list): return None
        if not path: return []
        
        if isinstance( path, str ): 
            return os.path.expandvars( path )
        
        for idx, p in enumerate(path):
            path[idx] = os.path.expandvars( p )
        
        return path

    @model_validator(mode="after")
    def default_options(self) -> 'RsyncCfg':
        if self.options is not None: return self
        self.options = RsyncOptions()
        return self
    
class ScheduleCfg(BaseModel):
    model_config = ConfigDict(extra="forbid", validate_default=True)

    weekday : CronField = None
    month   : CronField = None
    day     : CronField = None
    hour    : CronField = None
    minute  : CronField = None

    @field_validator("weekday", "month", "day", "hour", "minute", mode="before")
    @classmethod
    def normalize_fields(cls, field_value) -> str:
        """ Normalize fields to strings for cron validator """
        if field_value is None: return "*"
        if isinstance(field_value, int): return str(field_value)
        if isinstance(field_value, str) and field_value.strip():
            return field_value.strip()
        raise ValueError("must be null, int, or a cron field string")
    
class NotifType(str, enum.Enum):
    discord = "discord"

class EventType(str, enum.Enum):
    on_success = "success"
    on_failure = "failure"
    
class WebhookCfg(BaseModel):
    model_config = ConfigDict(extra="forbid", populate_by_name=True)

    name        : str
    url         : HttpUrl
    type_       : NotifType               = Field(alias="type")
    events      : List[EventType]         = Field(default_factory=list, min_length=1)
    timeout     : Optional[str]           = None
    max_retries : Optional[int]           = Field( default=None, ge=0 )
    headers     : Optional[Dict[str,Any]] = None

class SmtpCfg(BaseModel):
    model_config = ConfigDict(extra="forbid")

    server : str
    port   : Optional[int] = Field(default=None, ge=1, le=65535)
    ssl    : bool          = False

class EmailCfg(BaseModel):
    model_config = ConfigDict(extra="forbid", populated_by_name=True)

    from_    : EmailStr          = Field(alias="from")
    to       : List[EmailStr]    = Field(default_factory=list, min_length=1)
    password : str               = Field(str, min_length=1)
    smtp     : Optional[SmtpCfg] = None

class NotificationCfg(BaseModel):
    model_config = ConfigDict(extra="forbid")

    email    : Optional[EmailCfg]         = None
    webhooks : Optional[List[WebhookCfg]] = None

class LogRetentionCfg(BaseModel):
    max_spare_files  : int = Field( ge=1 )
    retention_window : int = Field( ge=1 )

class TargetCfg(BaseModel):
    model_config = ConfigDict(extra="forbid")
    
    remote        : RemoteCfg
    rsync         : RsyncCfg
    schedule      : ScheduleCfg
    notification  : Optional[NotificationCfg] = None
    log_retention : Optional[LogRetentionCfg] = LogRetentionCfg(
        max_spare_files=10, retention_window=7
    )

class NamedTarget(RequestModel, TargetCfg):
    """ Just a wrapper around target that also includes the name """
    name: str

    @classmethod
    def from_target(cls, name: str, target: TargetCfg) -> 'NamedTarget':
        return cls.model_construct( **target.__dict__, name=name )
    
class BackupCfg(BaseModel):
    model_config = ConfigDict(extra="forbid")
    targets: Optional[Dict[str, TargetCfg]] = None

class JobsConfiguration(BaseModel):
    model_config = ConfigDict(extra="forbid")
    backup: BackupCfg

ROOT_SCHEMA_CLASS = JobsConfiguration # For schema generation

class JobRunRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")
    
    dry_run : bool
    notify  : bool
    log     : bool