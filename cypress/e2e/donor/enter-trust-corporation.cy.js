import {TestEmail} from "../../support/e2e";

describe('Enter trust corporation', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/enter-trust-corporation');
    });

    it('can be submitted', () => {
        cy.checkA11yApp();

        cy.get('#f-name').invoke('val', 'Yoyodyne');
        cy.get('#f-email').invoke('val', TestEmail);

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/enter-trust-corporation-address');

        cy.contains("Add the trust corporation’s address");
    });

    it('errors when empty', () => {
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter trust corporation name');
        });

        cy.contains('[for=f-name] + .govuk-error-message', 'Enter trust corporation name');
    });

    it('errors when invalid email', () => {
        cy.get('#f-email').invoke('val', 'not-an-email');

        cy.contains('button', 'Save and continue').click();

        cy.contains('[for=f-email] + div + .govuk-error-message', 'Trust corporation email address must be in the correct format, like name@example.com');
    });
});
