# C4 Models via Structurizr

## Purpose

To provide an up to date Diagram as Code C4 Model of our application model to maintain a overview of the structure.

This should be updated as changes to the application occur and can be used to plan future iterations of the application structure.

## Instructions

The local folder contains Diagrams as Code that can be run with the following commands.

### To run in browser locally

- Run `make run-structurizr`
- Open a browser to [http://localhost:8080](http://localhost:8080)
- Open the workspace.dsl file and edit
- Every 2 seconds the website will check for updates and refresh if changes are detected

### To export Mermaid files

- Ensure you have installed `brew install structurizr-cli`
- Run `make run-structurizr-export`
- Commit and save the exported `.mmd` files
