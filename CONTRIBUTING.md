# Contribution guide

This repository is a library of Akeyless Custom Producer implementations.
Contributions that add new producers or extend/fix existing ones are welcome!

## New producers

To contribute a new producer, please follow the same structure/patterns used in
existing producers. Every producer must include a README.md file that documents
its deployment and usage.

## Existing producers

Some producers may benefit from new features. For example, `letsencrypt`
producer doesn't support many DNS providers, so adding new ones makes sense.

Bug fixes to existing producers are also welcome.

Another contribution path is adding a new kind of deployment. For example,
`echoserver` producer currently can be deployed only using AWS Lambda. Adding
other types of deployment could be a great example for future
users/contributors.

Please make sure to keep the changes backwards compatible.
