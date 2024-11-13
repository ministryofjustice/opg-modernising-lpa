# Managing node versions for development

To ensure we're using the right version of Nodejs locally, especially when switching between projects and repositories,
we can use [asdf](https://asdf-vm.com/).

These instructions are a summary of those found on asdf's [Getting Started guide](https://asdf-vm.com/guide/getting-started.html).

```shell
brew install asdf
asdf plugin add nodejs https://github.com/asdf-vm/asdf-nodejs.git
asdf install nodejs 18.7.0
asdf local nodejs 18.7.0
```

From here you can install Node dependencies with Yarn.

```shell
brew install yarn
```

If you encounter unrecognised versions of node after adding them to asdf you may need reinstall yarn for node re-shims to be recognised:

```shell
npm install -g yarn
```
