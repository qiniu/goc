// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';
import { GocServer } from './gocserver';

// this method is called when your extension is activated
// your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {

	let gocStatusBarItem: vscode.StatusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 0);
	gocStatusBarItem.text = 'Goc Coverage OFF';
	gocStatusBarItem.command = 'extension.switch';
	gocStatusBarItem.show();

	let gocserver = new GocServer();

	let disposable2 = vscode.commands.registerCommand('extension.switch', async () => {
		if (gocStatusBarItem.text == 'Goc Coverage OFF') {
			gocStatusBarItem.text = 'Goc Coverage ON';
			// get current project package structure
			let packages = gocserver.getGoList();
			await gocserver.startQueryLoop(packages);
		} else {
			gocStatusBarItem.text = 'Goc Coverage OFF';
			gocserver.stopQueryLoop();
		}
	})

	context.subscriptions.push(disposable2);
}

export function deactivate() {}
