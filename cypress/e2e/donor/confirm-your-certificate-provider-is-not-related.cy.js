describe('Confirm your certificate provider is not related', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/your-name&progress=addCorrespondent');
        cy.get('#f-last-name').clear().type('Cooper');
        cy.contains('button', 'Save and continue').click();
        cy.visitLpa('/task-list');
        cy.contains('li', "Check and send to your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();
    });

    it('continues when selected', () => {
        cy.checkA11yApp();

        cy.get('#f-yes-no').click({ force: true });
        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/check-your-lpa');
    });

    it('errors when not selected', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select the box to confirm your certificate provider is not related to you or your attorneys, or choose another certificate provider');
        });

        cy.contains('.govuk-error-message', 'Select the box to confirm your certificate provider is not related to you or your attorneys, or choose another certificate provider');
    });
});
