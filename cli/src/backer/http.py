"""
Module with some useful functions to communicate
with the backerd (the backer daemon) using its HTTP
REST API. 
"""
from __future__ import annotations

import os
import aiohttp
import aiohttp.typedefs
import backer.apitypes.response as resptypes
import backer.apitypes.requests as reqtypes

from enum import StrEnum
from pathlib import Path
from dataclasses import dataclass, field

from typing import (
    overload, Dict, Any, TypedDict, 
    Unpack, TypeVar, Type, Union
)

_T = TypeVar('_T')

SOCK_NAME = "backerd.sock"

class _RequestOptions(TypedDict, total=False):
    params  : aiohttp.typedefs.Query
    headers : Dict[str, Any]
    body    : Union[Dict[str,Any], reqtypes.RequestModel]

class HttpMethod(StrEnum):
    NULL = ""
    GET  = "GET"
    POST = "POST"

@dataclass(frozen=True)
class Endpoint:
    route  : str
    method : HttpMethod = HttpMethod.NULL
    models : Dict[int,Type[resptypes.ResponseModel]] = field(default_factory=dict)

    @overload
    def __truediv__(self, other: str) -> 'Endpoint': ...
    @overload
    def __truediv__(self, other: 'Endpoint') -> 'Endpoint': ...

    def __truediv__(self, other) -> 'Endpoint':
        if not isinstance(other, (str, Endpoint)):
            return NotImplemented

        if isinstance(other, str):
            # Input strings usually is for a route group, therefore
            # we can leave the model and method to default None, NULL
            return Endpoint(self.route + other)
        
        return Endpoint(self.route + other.route, other.method, other.models)
    
    def __eq__(self, other: object) -> bool:
        if not isinstance(other, Endpoint): return False
        return self.route  == other.route and self.method == other.method
    
    def get_model(self, code: int) -> Type[resptypes.ResponseModel]:
        return self.models.get(code, resptypes.BaseResponse)

class Endpoints:
    __Version_Models = {200: resptypes.VersionResponse}
    __Healthz_Models = {200: resptypes.HealthzResponse}

    __Jobs      = Endpoint("/jobs", HttpMethod.GET)
    __CreateJob = Endpoint("/create", HttpMethod.POST)
    __RunJob    = Endpoint("/run", HttpMethod.POST)

    _Root   = Endpoint("/api")
    Version = _Root / Endpoint("/version", HttpMethod.GET, __Version_Models)
    Healthz = _Root / Endpoint("/healthz", HttpMethod.GET, __Healthz_Models)

    def __init__(self): self._version = "/v1"
    def version(self, v: int) -> None:
        self._version = f"/v{v}"
    
    @property
    def _RootVers(self) -> Endpoint: return self._Root / self._version
    @property
    def Jobs(self) -> Endpoint: return self._RootVers / self.__Jobs
    @property
    def CreateJob(self) -> Endpoint: return self.Jobs / self.__CreateJob
    
    def RunJob(self, jn: str) -> Endpoint: return self.Jobs / jn / self.__RunJob

@dataclass(frozen=True)
class _Response[_T]:
    code : int
    body : _T

def get_unix_socket() -> str:
    """ Returns the unix socket for backerd """
    basedir = os.getenv("XDG_RUNTIME_DIR")
    return str(Path(basedir) / SOCK_NAME)

class BackerClient:
    def __init__(self, sock_path: str) -> None:
        self._sock_path = sock_path
        self._session: aiohttp.ClientSession | None = None

    async def make_request(
        self, endpoint: Endpoint, **kwargs: Unpack[_RequestOptions] 
    ) -> _Response[_T | Dict[str,Any]]:
        assert self._session is not None, "backer client must have a session"

        # Check the body type that must be either a dict or a subclass
        # or RequestModel. If it is not, then an error is raised. This must
        # be checked only if the body is not None
        request_body = kwargs.get('body', None)
        if request_body is not None and not isinstance(request_body, (dict, reqtypes.RequestModel)):
            raise ValueError(
                 "body object must be either a dict or a RequestModel derived class, " +
                f"got {type(kwargs['body'])} instead."
            )
        
        # If it is a derived class of RequestModel we need to get the dict out of it
        if isinstance(request_body, reqtypes.RequestModel):
            request_body = request_body.model_dump(mode="json", by_alias=True)

        # If body is not none then we need to set the Content-type to application/json
        if request_body is not None:
            curr_headers = kwargs.get("headers", dict())
            curr_headers['Content-Type'] = 'application/json'
            kwargs["headers"] = curr_headers

        # Reset the request body which is either unchanged or modified
        if "body"in kwargs: kwargs.pop('body')
        kwargs["json"] = request_body

        resp = await self._session.request(
            method=endpoint.method.value, url=f"http://localhost{endpoint.route}", **kwargs
        )
        
        data = await resp.json() # Get the json body from the response

        # If there are response models declared for this endpoints, we need
        # to pick the one matching the exit status (http response status)
        # and create the corresponding model
        data = endpoint.get_model(resp.status).model_validate(data)
        custom_resp = _Response(resp.status, data)
        resp.close() # Close the response before leaving
        return custom_resp

    async def create(self) -> None:
        connector = aiohttp.UnixConnector(path=self._sock_path)
        self._session = aiohttp.ClientSession(connector=connector)

    async def release(self) -> None:
        if self._session is None: return
        await self._session.close()
        self._session = None

    async def __aenter__(self) -> 'BackerClient':
        await self.create()
        return self
    
    async def __aexit__(self, *args):
        await self.release()