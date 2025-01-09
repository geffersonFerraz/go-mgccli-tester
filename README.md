# CLI Testing Library

This is a testing library designed for Command Line Interfaces (CLIs), with a primary focus on the CLI for **MagaluCloud**.

## Adding New Commands

To include new commands for testing, it is currently recommended to modify the `commands.yaml` file directly.

### [MGC Only]
For each command, if you do not specify an `api-key`, the test will use your current CLI authentication on your local machine.

---

## Running Tests

Executing tests is straightforward. Use the following command:

```bash
./go-mgccli-tester run -r=true -s=false
```

## Here is what each flag does:

    -r: Set this to true to run only the tests marked as read-only in the commands.yaml file.
    -s: Set this to false to ensure that tests do not overwrite snapshot files. A new snapshot will only be created if it doesn’t already exist.


## Recommended Workflow for Adding New Commands

    Add the new command to the commands.yaml file.
    Run the tests using the following command:

    ./go-mgccli-tester run -s=false

    This ensures that a new snapshot is created only if it doesn’t already exist.
    If the command is read-only, use the -r=true flag to restrict tests to read-only commands.

## By following this workflow, you can maintain a clean and efficient testing process.    