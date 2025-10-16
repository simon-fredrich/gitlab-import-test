# Gitlab Import Test
Fetch id of a gitlab project based on name and path of the project.
## Usage
You must be in the project root in order for it to work.

### Searching for Project
`go run . --namespaceId=<namespaceId> --path=<path>`

### Searching for Group
`go run . --parentId=<parentId> --path=<path>`