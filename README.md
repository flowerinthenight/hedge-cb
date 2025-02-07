[![main](https://github.com/flowerinthenight/hedge-cb/actions/workflows/main.yml/badge.svg)](https://github.com/flowerinthenight/hedge-cb/actions/workflows/main.yml)

## hedge-cb

An AWS-native cluster membership library. It is built on top of [aws/clock-bound](https://github.com/aws/clock-bound) (via [spindle-cb](https://github.com/flowerinthenight/spindle-cb)), making [CGO](https://pkg.go.dev/cmd/cgo) a requirement. It is a port (subset only) of [hedge](https://github.com/flowerinthenight/hedge). Included features from `hedge` include:

* Dynamic tracking of member nodes - good for clusters with members changing dynamically overtime, such as [Kubernetes Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/), [GCP Instance Groups](https://cloud.google.com/compute/docs/instance-groups), [AWS Autoscaling Groups](https://docs.aws.amazon.com/autoscaling/ec2/userguide/auto-scaling-groups.html), etc;
* Leader election - it maintains a single leader across the cluster;
* [Streaming] Send - all member nodes can send messages to the leader at any time;

A sample cloud-init [startup script](./startup-aws-asg.sh) is provided for spinning up an [Auto Scaling Group](https://docs.aws.amazon.com/autoscaling/ec2/userguide/auto-scaling-groups.html) with the ClockBound daemon already setup and running. You need to update the `ExecStart` section first with a working connection value. Note that this is NOT recommended though. You should use something like IAM Role + Secrets Manager, for instance.

```sh
# Create a launch template. ImageId here is Amazon Linux, default VPC.
# (Added newlines for readability. Might not run when copied as is.)
$ aws ec2 create-launch-template \
  --launch-template-name hedge-lt \
  --version-description version1 \
  --launch-template-data '
  {
    "UserData":"'"$(cat startup-aws-asg.sh | base64 -w 0)"'",
    "ImageId":"ami-0fb04413c9de69305",
    "InstanceType":"t2.micro",
  }'

# Create the single-zone ASG; update {target-zone} with actual value:
$ aws autoscaling create-auto-scaling-group \
  --auto-scaling-group-name hedge-asg \
  --launch-template LaunchTemplateName=spindle-lt,Version='1' \
  --min-size 3 \
  --max-size 3 \
  --tags Key=Name,Value=hedge-asg \
  --availability-zones {target-zone}

# or a multi-zone ASG; update {subnet(?)} with actual value(s):
$ aws autoscaling create-auto-scaling-group \
  --auto-scaling-group-name hedge-asg \
  --launch-template LaunchTemplateName=spindle-lt,Version='1' \
  --min-size 3 \
  --max-size 3 \
  --tags Key=Name,Value=hedge-asg \
  --vpc-zone-identifier "{subnet1,subnet2,subnet3}"


# You can now SSH to the instance. Note that it might take some time before
# ClockBound is running due to the need to build it in Rust. You can wait
# for the `clockbound` process, or tail the startup script output, like so:
$ tail -f /var/log/cloud-init-output.log

# Tail the service logs:
$ journalctl -f -u hedge
```
