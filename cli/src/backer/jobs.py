import argparse
import backer.http as httputils
import http
import pathlib as path
import yaml
import backer.apitypes.requests as reqtypes
import backer.apitypes.response as resptypes

from rich.progress import Progress, SpinnerColumn, TextColumn
from typing import TypeAlias

ResponseType: TypeAlias = httputils._Response[resptypes.BaseResponse]

async def handle_jobs_create(
    args: argparse.Namespace, client: httputils.BackerClient, endpoints: httputils.Endpoints
) -> None:
    """ Handle the `backer jobs create <config>` command """
    # First we need to take the configuration file from the parsed arguments and 
    # check that it exists and it is a YAML file
    configuration_file = path.Path(args.config).expanduser().absolute()
    if not configuration_file.exists():
        raise FileNotFoundError(f"{args.config} does not exists")
    
    if (
            not configuration_file.is_file()             \
        or  not str(configuration_file).endswith('.yml') \
        and not str(configuration_file).endswith('.yaml')
    ):
        raise ValueError(f"{args.config} must be a YAML file")
    
    # Once input argument validation has finished we need to parse the YAML 
    # file into the specific object
    data = yaml.safe_load(open(configuration_file, mode='r', encoding='utf-8'))
    configuration = reqtypes.JobsConfiguration.model_validate(data)

    # For each target sends a request to the daemon using the client
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        transient=True,
    ) as progress:
        task = progress.add_task("Starting ...", total=None)

        for name, target_cfg in configuration.backup.targets.items():
            named_target = reqtypes.NamedTarget.from_target(name, target_cfg)
            progress.update(task, description=f"Creating job [bold]{named_target.name}[/bold]…")

            resp: ResponseType = await client.make_request(endpoints.CreateJob, body=named_target)
            progress.console.print(f"([cyan]{resp.code}[/cyan]) {resp.body}")
        
        progress.update(task, description="[green]All jobs created[/green]")