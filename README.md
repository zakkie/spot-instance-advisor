# spot-instance-advisor

`spot-instance-advisor` is a tool that retrieves and displays AWS EC2 spot instance prices and interruption data by instance type.

## Installation

```
go install github.com/zakkie/spot-instance-advisor/cmd/spot-instance-advisor@latest
```

## Usage

```sh
spot-instance-advisor | sort -n -k 2,2 -r
```

another options:

```sh
spot-instance-advisor -help
Usage of ./bin/spot-instance-advisor:
  -max-memory int
        Memory in GB (default 32)
  -max-vcpus int
        Number of CPUs (default 4)
  -min-memory int
        Memory in GB (default 16)
  -min-vcpus int
        Number of CPUs (default 4)
  -region string
        AWS region (default "us-west-2")
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
