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

// Adds a table to the terminal with violation details
Cypress.Commands.add('checkA11yVvv', () => {
    cy.checkA11y(null, { rules: { region: { enabled: false } } }, terminalLog);
})

Cypress.Commands.add('addPersonToNotify', (p) => {
    cy.url().should('contain', '/choose-people-to-notify');

    cy.injectAxe();
    cy.checkA11y(null, { rules: { region: { enabled: false } } });

    cy.get('#f-first-names').type(p.firstNames)
    cy.get('#f-last-name').type(p.lastName)
    cy.get('#f-email').type(p.email)

    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/choose-people-to-notify-address');

    cy.injectAxe();
    cy.checkA11y(null, { rules: { region: { enabled: false } } });

    cy.get('#f-lookup-postcode').type(p.address.postcode)
    cy.contains('button', 'Find address').click();

    cy.url().should('contain', '/choose-people-to-notify-address');

    cy.injectAxe();
    cy.checkA11y(null, { rules: { region: { enabled: false } } });

    cy.get('#f-select-address').select(`${p.address.line1}, ${p.address.town}, ${p.address.postcode}`);
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/choose-people-to-notify-address');

    cy.injectAxe();
    cy.checkA11y(null, { rules: { region: { enabled: false } } });

    cy.get('#f-address-line-1').should('have.value', p.address.line1);
    cy.get('#f-address-line-2').should('have.value', p.address.line2);
    cy.get('#f-address-line-3').should('have.value', p.address.line3);
    cy.get('#f-address-town').should('have.value', p.address.town);
    cy.get('#f-address-postcode').should('have.value', p.address.postcode);

    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/choose-people-to-notify-summary');
})
