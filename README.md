# spot-instance-advisor

`spot-instance-advisor` is a tool that retrieves and displays AWS EC2 spot instance prices and interruption data by instance type.

## Installation

```
go install github.com/zakkie/spot-instance-advisor/cmd/spot-instance-advisor@latest
```

## Usage

```
spot-instance-advisor
```

### pre-conditions

Ensure the following pre-conditions are met before running the tool:

- You can use aws-cli
- You have the necessary permission about EC2
- The following command must succeed:
  ```
  aws ec2 describe-instance-types
  ```

## Developemnt

To set up the development environment, run the following commands:

```
make devel-deps
make run
```
