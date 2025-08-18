import {appendFileSync, readFileSync} from "fs";

const content = readFileSync(`${process.cwd()}/test-results/cypress-output.txt`, "utf8");
const regex = /cypress.+.cy.js(\s+│.+){3}\s+│.[^0]/g;
const matches = content.match(regex)

if(matches.length > 0) {
    let specs = '';

    matches.forEach((e) => { specs += ` ${e.substring(0, e.indexOf('.js')+3)}` });

    if (process.env.GITHUB_OUTPUT && specs !== "") {
        // for later use in actions
        appendFileSync(process.env.GITHUB_OUTPUT, `failedSpecs=${specs.trim()}\n`);
    }

    console.log(`Failed specs: ${specs.trim()}`);
}
