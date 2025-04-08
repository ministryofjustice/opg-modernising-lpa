describe('What country do you live in', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/what-country-do-you-live-in');
    });

    it('a11y', () => {
        cy.checkA11yApp();
    });

    context('when UK is selected', () => {
        beforeEach(() => {
            cy.get('#f-country').select('GB');
            cy.contains('button', 'Continue').click();
        });

        it('redirects to the UK address entry', () => {
            cy.url().should('include', '/your-address');
        });
    });

    context('when another country is selected', () => {
        beforeEach(() => {
            cy.get('#f-country').select('FR');
            cy.contains('button', 'Continue').click();
        });

        it('redirects to the non-UK address entry', () => {
            cy.url().should('include', '/your-non-uk-address');
        });
    });

    context('when unselected', () => {
        beforeEach(() => {
            cy.contains('button', 'Continue').click();
        })

        it('shows an error', () => {
            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Select the country you live in');
            });

            cy.contains('.govuk-label-wrapper + .govuk-error-message', 'Select the country you live in');
        });
    });
});
