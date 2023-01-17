# Contributing

We welcome contributions from the community. Please read the following guidelines carefully to
maximize the chances of your PR being merged.

## Communication

- If you want to work on some open issue, please reach out to us via a comment on it before starting your work to
	check if someone else is already working on it.
- Before working on a new feature please open an issue for the feature request and partecipate in the discussion for
	reaching an agreement on its design and utility for the project.
- Small patches, bug fixes and documentation typos fixing don’t need prior communication.

## Inclusive Language

Every PR, issue, code and documentation must be inclusive to all and must adhere to the following guidance:

- Every documentation should follow an inclusive style. A nice writeup has been done by google in its [Google Developer
	Documentation Style Guide].
- Every contribution will be covered by our [Code of Conduct](./CODE_OF_CONDUCT.md) so read it carefully.
- We will follow and will amend this list with the best practice and guidance that will emerge in the industry in the
	future and more comments and correction can be made during review by the mantainers.

## Opening a PR

- Fork the repo
- Read the [README.md](./README.md) file and the others documentation files, if presents, for guidances on how to setup
	the project locally and running the tests.
- If your PR is adding codes it must also contains test that will cover it, try to cover 100% of your added code if
	possibile. During the review be ready to explain why you cannot reach that percentage.
- We will not merge PR with failing tests or that will lower the coverage of the existing ones.
- When you open a PR please follow the indication that are provided in the template and provide all the relevant
	information
- Your PR title should be descriptive.
- If your PR is co-authored or based on an earlier PR from another contributor,
	please attribute them with `Co-authored-by: name <name@example.com>`.
	See [GitHub’s multiple author guidance] for further details.

## Commit Message Styling

Every commit in this repository must follow the guidelines provided by [Conventional commits].
The following *types* are allowed:

1. `fix:` a commit that fixes a bug.
1. `feat:` a commit that adds new functionality.
1. `docs:` a commit that adds or improves the documentation.
1. `test:` a commit that adds unit tests.
1. `ci:` a commit that improves the pipelines or the integration mechanisms.
1. `style:` a commit that changes the code or documentation format and/or style without modifying the implementation.
1. `chore:` a catch-all type for any other commits. Generally used for commits that do not add or improve
		functionalities to code or documentation.

[Google Developer Documentation Style Guide]: https://developers.google.com/style/inclusive-documentation
[GitHub’s multiple author guidance]: https://docs.github.com/en/pull-requests/committing-changes-to-your-project/creating-and-editing-commits/creating-a-commit-with-multiple-authors
[Conventional commits]: https://www.conventionalcommits.org/en/v1.0.0/
