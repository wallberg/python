# Fun with Programming

My playground for programming with Python 3.

## Python Environment

Set up the local Python environment:

```bash
# Get the project
git clone git@github.com:wallberg/sandbox.git
cd sandbox

# Setup the Python version
pyenv install --skip-existing $(cat .python-version)

# Setup the virtual environment
python -m venv .venv --prompt "$(basename "$PWD")-py$(cat .python-version)"
source .venv/bin/activate

# Install the requirements and setup any script entry points
pip install -r requirements.txt
```

## Extras

Project Euler code is tracked in a separate private repo, but can be included here:

` git clone git@github.com:wallberg/project-euler.git com/github/wallberg/euler `
