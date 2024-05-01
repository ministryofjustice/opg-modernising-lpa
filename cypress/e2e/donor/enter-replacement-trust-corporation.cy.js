import { TestEmail } from "../../support/e2e";

describe('Enter replacement trust corporation', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/enter-replacement-trust-corporation');
    });

    it('can be submitted', () => {
        cy.checkA11yApp();

        cy.get('#f-name').type('Yoyodyne');
        cy.get('#f-company-number').type('123456');
        cy.get('#f-email').type(TestEmail);

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/enter-replacement-trust-corporation-address');

        cy.contains("Add the trust corporation’s address");
    });

    it('errors when empty', () => {
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter company name');
            cy.contains('Enter company number');
        });

        cy.contains('[for=f-name] + .govuk-error-message', 'Enter company name');
        cy.contains('[for=f-company-number] + div + .govuk-error-message', 'Enter company number');
    });

    it('errors when invalid email', () => {
        cy.get('#f-email').type('not-an-email');

        cy.contains('button', 'Save and continue').click();

        cy.contains('[for=f-email] + div + .govuk-error-message', 'Company email address must be in the correct format, like name@example.com');
    });
});
