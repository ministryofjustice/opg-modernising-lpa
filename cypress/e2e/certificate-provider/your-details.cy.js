import {TestEmail, TestMobile} from "../../support/e2e";

describe('Your details', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provider-your-details&completeLpa=1&asCertificateProvider=1');
    });

    it('can be completed', () => {
        cy.checkA11yApp();

        cy.contains('Jamie Smith');
        cy.contains('Jessie Jones');

        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.get('#f-email').type(TestEmail);
        cy.get('#f-mobile').type(TestMobile);

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/certificate-provider-your-address');

        cy.checkA11yApp();
    });

    it('errors when all empty', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter mobile number');
            cy.contains('Enter email address');
            cy.contains('Enter date of birth');
        });

        cy.contains('[for=f-mobile] + .govuk-error-message', 'Enter mobile number');
        cy.contains('[for=f-email] + .govuk-error-message', 'Enter email address');
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Enter date of birth');
    });

    it('errors when invalid dates of birth', () => {
        cy.get('#f-date-of-birth').type('1');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must include a month and year');

        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('2222');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be in the past');

        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').clear().type('1990');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be a real date');
    });

    it('errors when not a UK mobile', () => {
        cy.get('#f-mobile').type('not a mobile');
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Mobile number must be a UK mobile number, like 07700 900 982 or +44 7700 900 982');
        });

        cy.contains('[for=f-mobile] + .govuk-error-message', 'Mobile number must be a UK mobile number, like 07700 900 982 or +44 7700 900 982');
    });

    it('errors when not an email', () => {
        cy.get('#f-email').type('not an email');
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Email address must be in the correct format, like name@example.com');
        });

        cy.contains('[for=f-email] + .govuk-error-message', 'Email address must be in the correct format, like name@example.com');
    });
});
