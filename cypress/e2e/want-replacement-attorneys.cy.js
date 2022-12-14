describe('Do you want replacement attorneys', () => {
    it('wants replacement attorneys - acting jointly', () => {
        cy.visit('/testing-start?redirect=/do-you-want-replacement-attorneys&howAttorneysAct=jointly');
        cy.injectAxe();

        cy.get('div.govuk-warning-text').should('contain', 'Replacement attorneys are an important backup when attorneys are appointed to act jointly.')

        cy.get('input[name="want"]').check('yes')

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');
    });

    it('wants replacement attorneys - acting jointly for some and severally for others', () => {
        cy.visit('/testing-start?redirect=/do-you-want-replacement-attorneys&howAttorneysAct=mixed');
        cy.injectAxe();

        cy.get('div.govuk-warning-text').should('contain', 'You appointed your attorneys to act jointly for some decisions, and jointly and severally for others.')

        cy.get('input[name="want"]').check('yes')

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');
    });

    it('wants replacement attorneys - acting jointly and severally', () => {
        cy.visit('/testing-start?redirect=/do-you-want-replacement-attorneys&howAttorneysAct=jointly-and-severally');
        cy.injectAxe();

        cy.get('div.govuk-warning-text').should('not.exist')

        cy.get('input[name="want"]').check('yes')

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');
    });

    it('does not want replacement attorneys', () => {
        cy.visit('/testing-start?redirect=/do-you-want-replacement-attorneys&howAttorneysAct=jointly-and-severally');
        cy.injectAxe();

        cy.get('div.govuk-warning-text').should('not.exist')

        cy.get('input[name="want"]').check('no')

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');

        cy.contains('a', 'Choose your replacement attorneys').parent().parent().contains('Completed')
    });
});
