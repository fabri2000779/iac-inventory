# IaC inventory

## This project was a PoC, is possible i dont continue it, but i leave it open if it can help someone.

To make this work you will need to deploy a read only AWS role, you can done this with a Stackset for the aws organization, you will also need to export you AWS credential of an account that can read the organization accounts.

The role name is hardcoded in the Controller:
```
arn:aws:iam::%s:role/ReadOnlyRole
```
