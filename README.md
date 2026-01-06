# Taiga CLI (Taiga Command-Line Interface for Automation)

A tool to automate task creation in Taiga project management tool.

These tool help you create tasks in Taiga from the command line or from a bulk input file, with support for custom fields like Activity Date, Start Time, and Total Time Spent.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.
See deployment section for notes on how to deploy the project on a live system.

### Setup

Create a .env file in the script directory with the following variables:
```
TAIGA_URL="https://YOUR.TAIGA.URL"
TAIGA_USER="YOUR_TAIGA_USERNAME"
TAIGA_PASSWORD="YOUR_TAIGA_PASSWORD"
PROJECT_SLUG="YOUR_TAIGA_PROJECT_SLUG"
```
Replace the values with your actual Taiga credentials and IDs.

#### Run
Create multiple tasks from a structured input file.
`./taiga-cli tasks/task.txt`

#### File Format
```
STORY_REF_ID
Task Subject | YYYY-MM-DD | HH:MM | Minutes
Another Task | YYYY-MM-DD | HH:MM | Minutes

```
- File should be in Unix format or using "LF" instead of "CRLF"
- First line: The ID of the user story where tasks will be created
- Following lines: Task data with fields separated by  `|`  character:
  - Task subject
  - Activity date (YYYY-MM-DD format)
  - Start time (HH:MM format)
  - Time spent (in minutes)
- Don't forget to add empty line in the buttom of file

#### Logs
All operations are logged in the  logs  directory:
- `error.log`: Contains error messages
- `created_tasks_YYYY-MM.log`: Lists tasks created in a specific year and month

#### Example   

##### Example Task File
```
597462
Daily Standup Meeting | 2025-06-02 | 09:30 | 30
Code Review Session | 2025-06-02 | 14:00 | 120

```
Run With:
`./taiga-cli tasks/task.txt`

### Prerequisites

Prequisites packages:
* Go (Go Programming Language)
* GoReleaser (Go Automated Binaries Build)
* Make (Automated Execution using Makefile)

Optional packages:
* Docker (Application Containerization)

### Deployment

#### **Using Container**

1) Install Docker CE based on the [manual documentation](https://docs.docker.com/desktop/)

2) Run the following command on your Terminal or PowerShell
```sh
docker run -it \
  -v tasks:/usr/app/taiga-cli/tasks \
  -v logs:/usr/app/taiga-cli/logs \
  --name taiga-cli \
  --rm dimaskiddo/taiga-cli:latest \
  taiga-cli tasks/task.txt
```

#### **Using Pre-Build Binaries**

1) Download Pre-Build Binaries from the [release page](https://github.com/dimaskiddo/taiga-cli/releases)

2) Extract the zipped file

3) Run the pre-build binary
```sh
# MacOS / Linux
chmod 755 taiga-cli
./taiga-cli tasks/task.txt

# Windows
# You can double click it or using PowerShell
.\taiga-cli.exe tasks/task.txt
```

#### **Build From Source**

Below is the instructions to make this source code running:

1) Create a Go Workspace directory and export it as the extended GOPATH directory
```sh
cd <your_go_workspace_directory>
export GOPATH=$GOPATH:"`pwd`"
```

2) Under the Go Workspace directory create a source directory
```sh
mkdir -p src/github.com/dimaskiddo/taiga-cli
```

3) Move to the created directory and pull codebase
```sh
cd src/github.com/dimaskiddo/taiga-cli
git clone -b master https://github.com/dimaskiddo/taiga-cli.git .
```

4) Run following command to pull vendor packages
```sh
make vendor
```

5) Until this step you already can run this code by using this command
```sh
make run
```

6) *(Optional)* Use following command to build this code into binary spesific platform
```sh
make build
```

7) *(Optional)* To make mass binaries distribution you can use following command
```sh
make release
```

### Running The Tests

Currently the test is not ready yet :)

## Built With

* [Go](https://golang.org/) - Go Programming Languange
* [GoReleaser](https://github.com/goreleaser/goreleaser) - Go Automated Binaries Build
* [Make](https://www.gnu.org/software/make/) - GNU Make Automated Execution
* [Docker](https://www.docker.com/) - Application Containerization

## Authors

* **Dimas Restu Hidayanto** - *Initial Work* - [DimasKiddo](https://github.com/dimaskiddo)

See also the list of [contributors](https://github.com/dimaskiddo/taiga-cli/contributors) who participated in this project

## Annotation

You can seek more information for the make command parameters in the [Makefile](https://github.com/dimaskiddo/taiga-cli/-/raw/master/Makefile)

## License

Copyright (C) 2026 Dimas Restu Hidayanto

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
