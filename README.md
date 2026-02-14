# 🔧 Backer

### A small backup control plane for self-hosted infrastructure

> Currently working on branch `develop/mils1` for Release [v0.1.0](https://github.com/lmriccardo/backer/milestone/2)

Backer is a modular backup orchestration system composed of:

- **`backer`** — Python CLI application
- **`backerd`** — Execution Daemon
- **Backer Desktop** - graphical control interface (*planned*)

It is designed for homelabs, NAS environments, and small private infrastructure. 

Up to now, under the hood, it uses [`rsync`](https://linux.die.net/man/1/rsync), a powerful 
command-line utility used to efficiently synchronize and transfer files between local and 
remote directories. It minimizes data transfer by using delta-transfer algorithm that only 
copies the specific portion of files that have changed.

> Since it relies on `rsync`, Backer is strongly Unix-driven, but for the future I am
> planning to introduce also Windows support either through WSL, or by rewriting
> an rsync-clone by myself (as good as possible).

## 🧱 Architecture

```
                ┌────────────────────┐
                │  Backer Desktop    │
                └──────────┬─────────┘
                           │
                           │
┌────────────┐      ┌──────▼──────┐
│   backer   │ ───► │   backerd   │
│    (CLI)   │      │  (daemon)   │
└────────────┘      └──────┬──────┘
                           │
                           ▼
                      Backup Jobs
                      Logs
                      Notifications
```

Both the CLI and the Desktop application communicate with the daemon
via the provided REST API, therefore they can be viewd as clients while 
the daemon owns the execution of *Jobs*.

In Backer terminology a *Job* is a scheduled backup task that runs
automatically. Each job is defined declaratively via a configuration
YAML file (which can contains multiple job definition).

```yaml
# yaml-language-server: $schema=config.schema.json

backup:
  targets:
    simple_backup:
      remote:
        host: nas.domain
        user: admin
        password_file: .rsync_pass
        dest:
          module: backup
          folder: home
      rsync:
        excludes:
          - **/node_modules/*
          - **/.cache/*
          - **/cache/*
          - **/*.tmp
        sources:
          - /home/
      notification:
        email:
          from: user.email@gmail.com
          to:
            - user.email@gmail.com
          password: password
```

The previous is a non-exhaustive example of a YAML configuration file
for a single job called `simple_backup` (also known as *Target*). For
a better overview of all possible configurations see the example file
[config-example.yml](config-example.yml). 

## 🚧 Roadmap

Planned, In progress and Ready features, as well as releases can be viewed in the [*Backer Project*](https://github.com/users/lmriccardo/projects/9)

## 📦 Installation

Coming Soon ...