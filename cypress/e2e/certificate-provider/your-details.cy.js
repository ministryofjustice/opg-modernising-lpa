import {TestEmail, TestMobile} from "../../support/e2e";

describe('Your details', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provider-your-details&completeLpa=1&asCertificateProvider=1');
    });

    it('can be completed', () => {
        cy.checkA11yApp();

        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.get('#f-mobile').type(TestMobile);

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/certificate-provider-your-address');

        cy.checkA11yApp();
    });

    it('errors when all empty', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter your UK mobile number');
            cy.contains('Enter your date of birth');
        });

        cy.contains('[for=f-mobile] ~ .govuk-error-message', 'Enter your UK mobile number');
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Enter your date of birth');
    });

    it('errors when invalid dates of birth', () => {
        cy.get('#f-date-of-birth').type('1');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Your date of birth must include a month and year');

        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('2222');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Your date of birth must be in the past');

        cy.get('#f-date-of-birth').type('not');
        cy.get('#f-date-of-birth-month').type('valid');
        cy.get('#f-date-of-birth-year').clear().type('values');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Enter a valid date of birth');
    });

    it('errors when not a UK mobile', () => {
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.get('#f-mobile').type('not a mobile');
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter a UK mobile number, like 07700 900 982 or +44 7700 900 982');
        });

        cy.contains('[for=f-mobile] ~ .govuk-error-message', 'Enter a UK mobile number, like 07700 900 982 or +44 7700 900 982');
    });

    it('errors when not over 18', () => {
        const lastYear = (new Date(new Date().setFullYear(new Date().getFullYear() - 1))).getFullYear()

        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type(lastYear.toString());
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('You’ve entered a date of birth that means you are under 18, you must be 18 to be a certificate provider');
        });

        cy.contains('#date-of-birth-hint + .govuk-error-message', 'You’ve entered a date of birth that means you are under 18, you must be 18 to be a certificate provider');
    });
});
