[![main](https://github.com/flowerinthenight/hedge-cb/actions/workflows/main.yml/badge.svg)](https://github.com/flowerinthenight/hedge-cb/actions/workflows/main.yml)

WIP: A port of [hedge](https://github.com/flowerinthenight/hedge) for AWS.

A sample cloud-init [startup script](./startup-aws-asg.sh) is provided for spinning up an [Auto Scaling Group](https://docs.aws.amazon.com/autoscaling/ec2/userguide/auto-scaling-groups.html) with the ClockBound daemon already setup and running.

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
