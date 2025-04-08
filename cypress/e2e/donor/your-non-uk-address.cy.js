describe('Your non-UK address', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/what-country-do-you-live-in');
        cy.get('#f-country').select('FR');
        cy.contains('button', 'Continue').click();
    });

    it('a11y', () => {
        cy.checkA11yApp();
    });

    context('when completed', () => {
        beforeEach(() => {
            cy.get('#f-apartmentNumber').type('123a');
            cy.get('#f-town').type('Cool town');
            cy.contains('button', 'Save and continue').click();
        });

        it('redirects to the next page', () => {
            cy.url().should('contain', '/receiving-updates-about-your-lpa');
        });
    });

    context('when changing country', () => {
        beforeEach(() => {
            cy.contains('a', 'Change').click();
        });

        it('redirects to the non-UK address entry', () => {
            cy.url().should('include', '/what-country-do-you-live-in');
        });
    });

    context('when nothing entered', () => {
        beforeEach(() => {
            cy.contains('button', 'Save and continue').click();
        })

        it('shows errors', () => {
            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Enter an apartment number, building number or building name');
                cy.contains('Enter town, suburb or city');
            });

            cy.contains('.govuk-fieldset .govuk-error-message', 'Enter an apartment number, building number or building name');
            cy.contains('.govuk-fieldset .govuk-error-message', 'Enter town, suburb or city');
        });
    });
});
