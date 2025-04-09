import { AddressFormAssertions } from "../../support/e2e";

describe('Do you live in the UK', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/do-you-live-in-the-uk');
    });

    it('a11y', () => {
        cy.checkA11yApp();
    });

    context('when yes selected', () => {
        beforeEach(() => {
            cy.contains('label', 'Yes').click();
            cy.contains('button', 'Continue').click();
        });

        it('goes to the UK address entry', () => {
            cy.url().should('include', '/your-address');
        });
    });

    context('when no selected', () => {
        beforeEach(() => {
            cy.contains('label', 'No').click();
            cy.contains('button', 'Continue').click();
        });

        it('goes to country selection', () => {
            cy.url().should('include', '/what-country-do-you-live-in');
        });
    });

    context('when unselected', () => {
        beforeEach(() => {
            cy.contains('button', 'Continue').click();
        })

        it('shows an error', () => {
            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Select yes if you live in the UK, the Channel Islands or the Isle of Man');
            });

            cy.contains('.govuk-fieldset .govuk-error-message', 'Select yes if you live in the UK, the Channel Islands or the Isle of Man');
        });
    });
});
