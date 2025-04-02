// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add('login', (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add('drag', { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add('dismiss', { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite('visit', (originalFn, url, options) => { ... })

import 'cypress-file-upload';

function terminalLog(violations) {
    cy.task(
        'log',
        `${violations.length} accessibility violation${
            violations.length === 1 ? '' : 's'
        } ${violations.length === 1 ? 'was' : 'were'} detected`
    )
    // pluck specific keys to keep the table readable
    const violationData = violations.map(
        ({ id, impact, description, nodes }) => ({
            id,
            impact,
            description,
            nodes: nodes.length
        })
    )

    cy.task('table', violationData)
}

// Sets base axe config and displays failures in table format
Cypress.Commands.add('checkA11yApp', (options= {}) => {
    const opts = {rules: { region: { enabled: false } } }
    opts.rules = {...opts.rules, ...options.rules}

    cy.injectAxe()
    cy.checkA11y(null, opts, terminalLog);
});

Cypress.Commands.add('visitLpa', (path, opts = {}) => {
    cy.url().then(u => cy.visit(u.split('/').slice(3, -1).join('/') + path, opts));
});

// Function to poll a page until element contains text or timeout occurs
Cypress.Commands.add('waitForTextByReloading', (selector, expectedText) => {
    const options = {
        timeout: 10000,
        interval: 500,
    };

    const startTime = Date.now();

    const checkTextAndReload = () => {
        return cy.get('body').then($body => {
            const $el = $body.find(selector);
            const found = $el.length > 0 && $el.text().includes(expectedText);

            if (found) {
                return;
            }

            if (Date.now() - startTime >= options.timeout) {
                throw new Error(`Timed out after ${options.timeout}ms waiting for "${selector}" to contain "${expectedText}"`);
            }

            cy.reload();
            cy.wait(options.interval);
            cy.then(checkTextAndReload);
        });
    };

    return cy.then(checkTextAndReload);
});
