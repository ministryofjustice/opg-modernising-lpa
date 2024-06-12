describe('Life sustaining treatment', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/life-sustaining-treatment&lpa-type=personal-welfare&progress=chooseYourAttorneys');
    });

    it('can be agreed to', () => {
        cy.checkA11yApp();

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/task-list');
        cy.visitLpa('/restrictions');
        cy.contains('life-sustaining treatment');
    });

    it('can be disagreed with', () => {
        cy.checkA11yApp();

        cy.contains('label', 'No').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/task-list');
        cy.visitLpa('/restrictions');
        cy.contains('life-sustaining treatment').should('not.exist');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select if you do or do not give your attorneys authority to give or refuse consent to life-sustaining treatment on your behalf');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select if you do or do not give your attorneys authority to give or refuse consent to life-sustaining treatment on your behalf');
    });
});
