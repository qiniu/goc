# Goc Coverage

This extension provides rich support for the [goc](https://github.com/qiniu/goc) tool.

## Overview

* [Getting started](#getting-started)
* [Ask for help](#ask-for-help)

## Getting started

Welcome! The [goc](https://github.com/qiniu/goc) is a coverage tool for Golang projects. The most interesting part of [goc](https://github.com/qiniu/goc) is that it can generate coverage report while the service is running! You can test the service manually or automatically, whatever, you don't have to stop the tested service to get the coverage report anymore.

This extension provides a frontend to show the covered lines in real time.

### Basic requirements

Before you started, make sure that you have:

1. Go
2. [goc](https://github.com/qiniu/goc)
3. source code of the tested service

### Set up your environment

Follow the [goc example](https://github.com/qiniu/goc#examples) guide to build and start the tested service. After you finished this step, there should be a **goc server** running at default port `7777`.

Use `vscode` to open the source code. Make sure the vscode's `workspace` is in the Golang project's root directory. If your project uses `go module`, just open vscode in the repo's directory. If your project uses `GOPATH`, you should setup right `GOPATH` before you open vscode.

Open any Go source files, you should see a `Goc Coverage OFF` button in the bottom status bar, click on the button to enable rendering covered lines in real time.

### Configuration

#### goc server url

If you deploy the goc server on another host with a customized port, you can set:

```
"goc.serverUrl": "http://192.168.1.3:51234"
```

## Ask for help

If you're having issues with this extension, please reach out to us by [filing an issue](https://github.com/qiniu/goc/issues/new/choose) directly.