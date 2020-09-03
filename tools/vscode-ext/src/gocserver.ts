import axios from 'axios';
axios.defaults.timeout = 3000;
import * as vscode from 'vscode';
import { spawnSync } from 'child_process';
import * as path from 'path';
import { promisify } from 'util';
import * as log4js from 'log4js';
const sleep = promisify(setTimeout);

export class GocServer {
    private _serverUrl: string = '';
    private _debug: Boolean = false;
    private _logger: log4js.Logger = log4js.getLogger();
    private timer = true;
    private highlightDecorationType = vscode.window.createTextEditorDecorationType({
        backgroundColor: 'green',
        border:  '2px solid white',
        color:  'white'
    });;
    private lastProfile = '';
    private lastFileNeedsRender = '';

    construct() { }

    async startQueryLoop(packages: any[]) {
        this.timer = true;
        while ( true === this.timer ) {
            await sleep(2000);

            if ( true !== this.timer ) {
                this.clearHightlight();
                return;
            }

            this.getConfigurations();
            this.setDebugLogger();
            let profile = await this.getLatestProfile();
            this.renderFile(packages, profile);
        }
    }

    stopQueryLoop() {
        this.timer = false;
        this.lastProfile = '';
        this.lastFileNeedsRender = '';

        this.clearHightlight()
    }

    clearHightlight() {
        vscode.window.visibleTextEditors.forEach(visibleEditor => {
            visibleEditor.setDecorations(this.highlightDecorationType, []);
        });
    }

    getConfigurations() {
        this._serverUrl = vscode.workspace.getConfiguration().get('goc.serverUrl') || '';
        this._debug = vscode.workspace.getConfiguration().get('goc.debug') || false;
    }

    setDebugLogger() {
        if (this._debug == false) {
            this._logger.level = "info";
        } else {
            this._logger.level = "debug";
        }

        this._logger.info('goc server url: ', this._serverUrl);
    }

    async getLatestProfile(): Promise<string> {
        let profileApi = `${this._serverUrl}/v1/cover/profile?force=true`;

        try {
            let res = await axios.get(profileApi, );
            let body: string = res.data.toString();
            this._logger.debug(body);
            return body;
        } catch(err) {
            this._logger.error(err.message);
        } 

        return "";
    }

    checkGoEnv() : Boolean {
        let output = spawnSync('go', ['version']);
        if (output.status != 0 || output.status == null) {
            this._logger.error(output.stderr.toString());
            return true;
        }
        this._logger.debug('go version: ', output.stdout.toString());
        return false;
    }

    getGoList(): Array<any> {
        let cwd = "";
        let workspaces = vscode.workspace.workspaceFolders || [];
        if (workspaces.length == 0) {
            this._logger.error("no workspace found");
            return [];
        } else {
            cwd = workspaces[0].uri.path;
        }
        this._logger.debug('current project root directory: ', cwd);
        let opts = {
            'cwd': cwd
        };
        let output = spawnSync('go', ['list', '-json', './...'], opts);
        if (output.status != 0 || output.status == null) {
            this._logger.error(output.stderr.toString());
            return [];
        } 
        let packages = JSON.parse('[' + output.stdout.toString().replace(/}\n{/g, '},\n{') + ']');
        return packages;
    }

    renderFile(packages: Array<any>, profile: string) {
        let activeTextEditor = vscode.window.activeTextEditor;
        let fileNeedsRender = activeTextEditor?.document.fileName || '---';

        this._logger.debug('current active source code file: ', fileNeedsRender);
        // check if needs to rerender
        if (profile == this.lastProfile && fileNeedsRender == this.lastFileNeedsRender) {
            return;
        }
        this.lastProfile = profile;
        this.lastFileNeedsRender = fileNeedsRender;

        this._logger.debug('go list packages length: ', packages.length);

        for (let i=0; i<packages.length; i++) {
            let p = packages[i];
            let baseDir: string = p['Dir'];
            for (let gofile of p['GoFiles']) {
                let filepath = path.join(baseDir, gofile);
                // this._logger.debug('compare two path: ', filepath, fileNeedsRender);
                if (filepath == fileNeedsRender) {
                    let importPath: string = path.join(p['ImportPath'], gofile);
                    let ranges = this.parseProfile(profile, importPath);
                    this.triggerUpdateDecoration(ranges);
                    return;
                }
            }
        }
    }

    parseProfile(profile: string, importPathNeedsRender: string): vscode.Range[] {
        let lines = profile.split('\n');
        this._logger.debug('profile lines: ', lines.length);
        if (lines.length <= 1) {
            this._logger.error("empty coverage profile from server");
            return [];
        }


        let ranges: vscode.Range[] = [];
        for (let i=0; i<lines.length; i++) {
            let line = lines[i];
            let importPath: string = line.split(':')[0];
            let blockInfo: string = line.split(':')[1];
            
            if (importPath != importPathNeedsRender) {
                continue;
            }

            let rxp = /(\d+)\.(\d+),(\d+)\.(\d+)\s(\d+)\s(\d+)/g
            let matches = rxp.exec(blockInfo)!;
            let startLine = matches[1];
            let startCol = matches[2];
            let endLine = matches[3];
            let endCol = matches[4];
            let stmts = matches[5];
            let count = matches[6];

            // no need to render code block not covered
            if (count == '0') {
                continue;
            }

            let range = new vscode.Range(
                new vscode.Position(Number(startLine)-1, Number(startCol)-1),
                new vscode.Position(Number(endLine)-1, Number(endCol)-1)
            );
            
            ranges.push(range);
        }

        return ranges
    }

    triggerUpdateDecoration(ranges: vscode.Range[]) {
        if (!vscode.window.activeTextEditor) {
            return;
        }
      
        this._logger.info('[' + new Date().toUTCString() + '] ' + 'update latest profile success')

        if (ranges.length == 0) {
            this.clearHightlight();
        } else {
            vscode.window.activeTextEditor.setDecorations(
                this.highlightDecorationType,
                ranges
            )
        }
    }
}