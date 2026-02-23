from __future__ import annotations

import aiohttp

from packaging.version import Version
from datetime import datetime
from dataclasses import dataclass
from typing import List, Tuple, TypeAlias
from backer import __version__ as __in_version__
from importlib.metadata import version, PackageNotFoundError

__version__ = __in_version__
try:
    __version__ = version("backer-cli")
except PackageNotFoundError:
    ...

RELEASE_API_URL = "https://pypi.org/simple/backer-cli/"

@dataclass
class RemoteFileInfo:
    name: str
    upload_time: datetime

VersionList : TypeAlias = List[Version]
FileList    : TypeAlias = List[RemoteFileInfo]

async def _get_all_versions() -> Tuple[VersionList, FileList] | None:
    """ Get the lastest version of the backupctl project """
    headers = {'Host': 'pypi.org', 'Accept': 'application/vnd.pypi.simple.v1+json'}
    timeout = aiohttp.ClientTimeout(total=10)

    # Perform the async fetch operation and get back the payload
    async with aiohttp.ClientSession(timeout=timeout, headers=headers) as session:
        async with session.get(RELEASE_API_URL) as resp:
            if resp.status < 200 or resp.status >= 300: return 
            payload = await resp.json()

    if not isinstance(payload, dict): return
    if not "versions" in payload: return
    if not isinstance(payload["versions"], list): return
    if not "files" in payload: return

    # First parse all files
    version_list = list(map(Version, payload.get('versions')))
    file_list = []
    for file_data in payload.get("files"):
        if not isinstance(file_data, dict): continue

        file_name = file_data.get('filename')
        upload_time = file_data.get('upload-time')

        if not (isinstance(file_name, str) and isinstance(upload_time, str)):
            continue

        upload_time = datetime.fromisoformat(upload_time.replace("Z", "+00:00"))
        file_list.append(RemoteFileInfo( file_name, upload_time ))
        
    return version_list, file_list

def _get_latest_release( versions: VersionList ) -> Version:
    """ Returns the lastest release """
    return max( versions )

def _get_release_time( version: Version, files: FileList ) -> datetime | None:
    """ Return the date when the version has been released """
    version_str = str(version) # Convert back to string
    for fileitem in files:
        if fileitem.name.startswith(f"backer-{version_str}"):
            return fileitem.upload_time

    return None

async def format_version() -> None:
    """ Format the version and get latest version """
    curr_version = Version(__version__)
    result = await _get_all_versions()
    curr_version_time = None

    if result is not None:
        versions, files = result
        curr_version_time = _get_release_time( curr_version, files )

    # Print the current version
    print(f"CLI Version: {curr_version} (", end="")
    curr_t = "Not Yet Released" if curr_version_time is None else str(curr_version_time)
    print(f"{curr_t})")

    if result is not None:
        last_version = _get_latest_release( versions )
        last_version_time = _get_release_time( last_version, files )

        # Print if there is a more recent version
        if curr_version < last_version:
            print(
                "!! A new version is available - " +\
                f"{last_version} ({last_version_time}) !!"
            )
            
    print("")