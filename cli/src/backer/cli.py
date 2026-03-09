import argparse
import asyncio
import http
import aiohttp.client_exceptions as httpexcept
import backer.jobs as jobsfn
import backer.http as httputils
import backer.version as version

CLIENT : httputils.BackerClient | None = None

class UnreachableDaemonError(Exception):
    def __init__(self) -> None: super().__init__()
    def __str__(self) -> str:
        return "Backer daemon is not reachable or it is not running"
    
async def _exec(parser: argparse.ArgumentParser) -> None:
    # Parse the arguments
    args = parser.parse_args()

    global CLIENT
    CLIENT = httputils.BackerClient(sock_path=httputils.get_unix_socket())
    await CLIENT.create() # Create the session with the connector

    # Before performing any command, check if the daemon
    # is reachable by performing a simple healtz request.
    resp = await CLIENT.make_request(httputils.Endpoints.Healthz)
    if resp.code != http.HTTPStatus.OK or not resp.body.ok:
        raise UnreachableDaemonError()
    
    # In the health message there is also the api version
    endpoints = httputils.Endpoints()
    endpoints.version(resp.body.version.api)

    # If the --version option is given just returns the version
    if args.version:
        # Format the version and returns
        print(resp.body.version)
        await version.format_version()
        return

    # If the version is not given, but there is no
    # command as well, just prints out the help msg
    if args.command is None:
        parser.print_help()
        return 1
    
    _ = await args.func(args, CLIENT, endpoints)
    return 0

def def_jobs_command(parser: argparse._SubParsersAction) -> None:
    # Top-level subparser
    jobs_parser = parser.add_parser("jobs", help="Manage jobs")
    jobs_subparser = jobs_parser.add_subparsers(dest="jobs_command", required=True)

    # Jobs create command
    create_parser = jobs_subparser.add_parser("create", help="Create a job")
    create_parser.add_argument("config", help="YAML configuration file")
    create_parser.set_defaults(func=jobsfn.handle_jobs_create)

    # Jobs run command
    run_parser = jobs_subparser.add_parser("run", help="Run a job")
    run_parser.add_argument("job_name", help="The name of the job to run")
    run_parser.add_argument("--no-wait", help="Does not block the caller for run completition", action="store_true")
    run_parser.add_argument("--dry-run", help="Enable dry run mode", action="store_true")
    run_parser.add_argument("--notify", help="Enable notification", action="store_true")
    run_parser.add_argument("--log", help="Enable logging", action="store_true")

async def async_main() -> int:
    parser = argparse.ArgumentParser(prog="backer", description="Backup automation control tool")
    subparsers = parser.add_subparsers(dest="command")

    # First add --version option to base parser
    parser.add_argument("--version", required=False, action="store_true", help="Print out the version")

    def_jobs_command(subparsers)

    exit_code = 1

    try:
        exit_code = await _exec(parser)
    except (httpexcept.ClientConnectorError, UnreachableDaemonError):
        print(UnreachableDaemonError())
    except Exception as e:
        print(f"Error: {e} (type: {type(e)})")

    await CLIENT.release()
    return exit_code

def main() -> int:
    return asyncio.run(async_main())