describe('Can you sign your LPA', () => {
    describe('first time', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/can-you-sign-your-lpa');
        });

        it('can be submitted', () => {
            cy.checkA11yApp();
            cy.contains('a', 'Return to task list').should('not.exist');

            cy.get('#f-selected').check({ force: true });

            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-preferred-language');
        });

        it('errors when empty', () => {
            cy.contains('button', 'Save and continue').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Select yes if you can sign yourself online');
            });

            cy.contains('fieldset .govuk-error-message', 'Select yes if you can sign yourself online');
        });
    });

    describe('after completing', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/can-you-sign-your-lpa&progress=chooseYourAttorneys');
        });

        it('shows task list button', () => {
            cy.contains('a', 'Return to task list');
        });
    });
});
