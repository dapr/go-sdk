# Contribution Guidelines

Thank you for your interest in Dapr Go SDK!

This project welcomes contributions and suggestions. Most contributions require you to agree to a Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us the rights to use your contribution.

For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions provided by the bot. You will only need to do this once across all repositories using our CLA.

This project has adopted the Microsoft Open Source Code of Conduct. For more information see the Code of Conduct FAQ or contact opencode@microsoft.com with any additional questions or comments.

Contributions come in many forms: submitting issues, writing code, participating in discussions and community calls.

This document provides the guidelines for how to contribute to the Dapr Go SDK project.

## Issues

This section describes the guidelines for submitting issues

### Issue Types

There are 4 types of issues:

- Issue/Bug: You've found a bug with the code, and want to report it, or create an issue to track the bug.
- Issue/Discussion: You have something on your mind, which requires input from others in a discussion, before it eventually manifests as a proposal.
- Issue/Proposal: Used for items that propose a new idea or functionality. This allows feedback from others before code is written.
- Issue/Question: Use this issue type, if you need help or have a question.

### Before You File

Before you file an issue, make sure you've checked the following:

1. Check for existing issues
    - Before you create a new issue, please do a search in [open issues](https://github.com/dapr/go-sdk/issues) to see if the issue or feature request has already been filed.
    - If you find your issue already exists, make relevant comments and add your [reaction](https://github.com/blog/2119-add-reaction-to-pull-requests-issues-and-comments). Use a reaction:
        - 👍 up-vote
        - 👎 down-vote
1. For proposals
    - Some changes to the Dapr Go SDK may require changes to the API. In that case, the best place to discuss the potential feature is the main [Dapr repo](https://github.com/dapr/dapr).
    - Other examples could include bindings, state stores or entirely new components.

## Contributing to Dapr Go SDK

This section describes the guidelines for contributing code/docs to Dapr Go SDK.

### Pull Requests

All contributions come through pull requests. To submit a proposed change, we recommend following this workflow:

1. Make sure there's an issue (bug or proposal) raised, which sets the expectations for the contribution you are about to make.
1. Fork the relevant repo and create a new branch
1. Create your change
  - Code changes require tests
1. Update relevant documentation for the change
1. Run through pre-commit steps until everything passes 
  - `make test`
  - `make lint`
  - `make spell` 
1. Commit and open a PR
1. Wait for the CI process to finish and make sure all checks are green (including the test coverage)
1. A maintainer of the project will be assigned, and you can expect a review within a few days

#### Use work-in-progress PRs for early feedback

A good way to communicate before investing too much time is to create a "Work-in-progress" PR and share it with your reviewers. The standard way of doing this is to add a "[WIP]" prefix in your PR's title and assign the **do-not-merge** label. This will let people looking at your PR know that it is not well baked yet.

### Use of Third-party code

- All third-party code must be placed in the `vendor/` folder.
- `vendor/` folder is managed by Go modules which stores the source code of third-party Go dependencies. - The `vendor/` folder should not be modified manually.
- Third-party code must include licenses.

A non-exclusive list of code that must be places in `vendor/`:

- Open source, free software, or commercially-licensed code.
- Tools or libraries or protocols that are open source, free software, or commercially licensed.

**Thank You!** - Your contributions to open source, large or small, make projects like this possible. Thank you for taking the time to contribute.

## Code of Conduct

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
