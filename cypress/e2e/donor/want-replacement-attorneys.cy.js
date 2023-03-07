describe('Do you want replacement attorneys', () => {
    it('wants replacement attorneys - acting jointly', () => {
        cy.visit('/testing-start?redirect=/do-you-want-replacement-attorneys&howAttorneysAct=jointly');

        cy.get('div.govuk-warning-text').should('contain', 'Replacement attorneys are an important backup when attorneys are appointed to act jointly.')

        cy.get('input[name="want"]').check('yes')

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');
    });

    it('wants replacement attorneys - acting jointly for some and severally for others', () => {
        cy.visit('/testing-start?redirect=/do-you-want-replacement-attorneys&howAttorneysAct=mixed');

        cy.get('div.govuk-warning-text').should('contain', 'You appointed your attorneys to act jointly for some decisions, and jointly and severally for others.')

        cy.get('input[name="want"]').check('yes')

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');
    });

    it('wants replacement attorneys - acting jointly and severally', () => {
        cy.visit('/testing-start?redirect=/do-you-want-replacement-attorneys&howAttorneysAct=jointly-and-severally');

        cy.get('div.govuk-warning-text').should('not.exist')

        cy.get('input[name="want"]').check('yes')

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');
    });

    it('does not want replacement attorneys', () => {
        cy.visit('/testing-start?redirect=/do-you-want-replacement-attorneys&howAttorneysAct=jointly-and-severally');

        cy.get('div.govuk-warning-text').should('not.exist')

        cy.get('input[name="want"]').check('no')

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');

        cy.contains('a', 'Choose your replacement attorneys').parent().parent().contains('Completed')
    });

    it('errors when unselected', () => {
        cy.visit('/testing-start?redirect=/do-you-want-replacement-attorneys&howAttorneysAct=jointly-and-severally');
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select yes to add replacement attorneys');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select yes to add replacement attorneys');
    });
});
