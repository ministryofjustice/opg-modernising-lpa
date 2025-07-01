const { randomAccessCode } = require("../../support/e2e");

describe('Enter access code', () => {
    let accessCode = ''
    beforeEach(() => {
        accessCode = randomAccessCode()

        cy.visit(`/fixtures/voucher?redirect=&withAccessCode=${accessCode}`);

        cy.contains('a', 'Start').click()
        cy.origin('http://localhost:7012', () => {
            cy.contains('button', 'Continue').click();
        });
        cy.url().should('contain', '/voucher-enter-access-code')
    });

    it('can enter a valid access code', { pageLoadTimeout: 6000 }, () => {
        cy.checkA11yApp();

        cy.get('#f-donor-last-name').type('Smith');
        cy.get('#f-access-code').invoke('val', accessCode);
        cy.contains('Save and continue').click();

        cy.url().should('contain', '/task-list')
    });

    it('errors when empty', () => {
        cy.contains('Save and continue').click();

        cy.checkA11yApp();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter donor’s last name');
            cy.contains('Enter your access code');
        });

        cy.contains('[for=f-donor-last-name] ~ .govuk-error-message', 'Enter donor’s last name');
        cy.contains('[for=f-access-code] ~ .govuk-error-message', 'Enter your access code');
    });

    it('errors when incorrect code', () => {
        cy.get('#f-access-code').invoke('val', 'wrongish');
        cy.get('#f-donor-last-name').type('What');
        cy.contains('Save and continue').click();

        cy.checkA11yApp();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The access code you entered is incorrect, please check it and try again');
        });

        cy.contains('[for=f-access-code] ~ .govuk-error-message', 'The access code you entered is incorrect, please check it and try again');
    });

    it('errors when incorrect code length', () => {
        cy.get('#f-access-code').invoke('val', 'short');
        cy.contains('Save and continue').click();

        cy.checkA11yApp();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The access code you enter must be 8 characters');
        });

        cy.contains('[for=f-access-code] ~ .govuk-error-message', 'The access code you enter must be 8 characters');
    });

});
