import { DateOfBirthAssertions } from "../../support/e2e";

describe('Enter date of birth', () => {
    it('can be completed', () => {
        cy.visit('/fixtures/certificate-provider?redirect=/enter-date-of-birth');
        cy.checkA11yApp();

        cy.get('#f-date-of-birth').invoke('val', '1');
        cy.get('#f-date-of-birth-month').invoke('val', '2');
        cy.get('#f-date-of-birth-year').invoke('val', '1990');

        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/your-preferred-language');
    });

    describe('errors and warnings', () => {
        beforeEach(() => {
            cy.visit('/fixtures/certificate-provider?redirect=/enter-date-of-birth');
        });

        it('can be over 100', () => {
            cy.checkA11yApp();

            cy.get('#f-date-of-birth').invoke('val', '1');
            cy.get('#f-date-of-birth-month').invoke('val', '2');
            cy.get('#f-date-of-birth-year').invoke('val', '1900');

            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/enter-date-of-birth');
            cy.contains('By continuing, you confirm that this person is more than 100 years old. If not, please change their date of birth.')

            cy.checkA11yApp();

            cy.contains('button', 'Save and continue').click();

            cy.url().should('contain', '/your-preferred-language');
        });

        it('errors when empty', () => {
            cy.contains('button', 'Save and continue').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Enter date of birth');
            });

            cy.contains('#date-of-birth-hint + .govuk-error-message', 'Enter date of birth');
        });

        it('errors when invalid dates of birth', () => {
            DateOfBirthAssertions.assertInvalidDatesOfBirth()
        });

        it('errors when not over 18', () => {
            const lastYear = (new Date(new Date().setFullYear(new Date().getFullYear() - 1))).getFullYear()

            cy.get('#f-date-of-birth').invoke('val', '1');
            cy.get('#f-date-of-birth-month').invoke('val', '2');
            cy.get('#f-date-of-birth-year').invoke('val', lastYear.toString());
            cy.contains('button', 'Save and continue').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('You’ve entered a date of birth that means you are under 18, you must be 18 to be a certificate provider');
            });

            cy.contains('#date-of-birth-hint + .govuk-error-message', 'You’ve entered a date of birth that means you are under 18, you must be 18 to be a certificate provider');
        });
    })
});
