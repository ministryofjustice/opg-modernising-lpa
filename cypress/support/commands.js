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

Cypress.Commands.add('setUploadsClean', (opts = {}) => {
    cy.url().then(u => cy.exec(`(make set-uploads-clean lpaId=${getLpaIDFromURL(u)})`));
});

Cypress.Commands.add('setUploadsInfected', (opts = {}) => {
    cy.url().then(u => cy.exec(`(make set-uploads-infected lpaId=${getLpaIDFromURL(u)})`));
});

function getLpaIDFromURL(url) {
    return url.split('/').slice(4, -1)[0]
}
