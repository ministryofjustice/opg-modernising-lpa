describe('Check you can sign', () => {
    describe('first time', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/check-you-can-sign');
        });

        it('can be submitted', () => {
            cy.checkA11yApp();
            cy.contains('a', 'Return to task list').should('not.exist');

            cy.get('#f-yes-no').check({ force: true });

            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-preferred-language');
        });

        it('errors when empty', () => {
            cy.contains('button', 'Save and continue').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Select yes if you will be able to sign the LPA yourself');
            });

            cy.contains('fieldset .govuk-error-message', 'Select yes if you will be able to sign the LPA yourself');
        });
    });

    describe('after completing', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/check-you-can-sign&progress=chooseYourAttorneys');
        });

        it('shows task list button', () => {
            cy.contains('a', 'Return to task list');
        });
    });
});
