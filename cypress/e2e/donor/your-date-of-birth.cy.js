import {DateOfBirthAssertions} from "../../support/e2e";

describe('Your date of birth', () => {
    describe('first time', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/your-date-of-birth');
        });

        it('can be submitted', () => {
            cy.checkA11yApp();
            cy.contains('a', 'Return to task list').should('not.exist');

            cy.get('#f-date-of-birth').invoke('val', '1');
            cy.get('#f-date-of-birth-month').invoke('val', '2');
            cy.get('#f-date-of-birth-year').invoke('val', '1990');

            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/do-you-live-in-the-uk');
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

        it.only('warns when date of birth is over 100', () => {
            cy.get('#f-date-of-birth').invoke('val', '1');
            cy.get('#f-date-of-birth-month').invoke('val', '2');
            cy.get('#f-date-of-birth-year').invoke('val', '1900');
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/warning');

            cy.contains('You are over 100 years old.');

            cy.contains('a', 'Continue').click();
            cy.url().should('contain', '/do-you-live-in-the-uk');
        });
    });

    describe('after completing', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/your-date-of-birth&progress=chooseYourAttorneys');
        });

        it('shows task list button', () => {
            cy.contains('a', 'Return to task list');
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/do-you-live-in-the-uk');
        });
    });
});
