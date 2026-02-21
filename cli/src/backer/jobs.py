import argparse
import backer.http as httputils

async def handle_jobs_create(
    args: argparse.Namespace, client: httputils.BackerClient, 
    endpoints: httputils.Endpoints
) -> None:
    """ Handle the `backer jobs create <config>` command """
    print("Create job called")