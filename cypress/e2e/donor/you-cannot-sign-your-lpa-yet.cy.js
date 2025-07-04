describe('You cannot sign your LPA yet', () => {
    it('lists attorneys and replacement attorneys that are under 18', () => {
        const today = new Date()
        cy.visit('/fixtures?redirect=/choose-attorneys-summary&progress=addCorrespondent');

        cy.contains('.govuk-summary-card', 'Jessie Jones').contains('a', 'Change').click();
        cy.get('#f-date-of-birth-year').invoke('val', today.getFullYear() - 1);
        cy.contains('button', 'Save and continue').click()
        cy.contains('a', 'Continue').click()

        cy.visitLpa('/choose-replacement-attorneys-summary')
        cy.contains('.govuk-summary-card', 'Blake Buckley').contains('a', 'Change').click();
        cy.get('#f-date-of-birth-year').invoke('val', today.getFullYear() - 1);
        cy.contains('button', 'Save and continue').click()
        cy.contains('a', 'Continue').click()
        cy.contains('button', 'Save and continue').click()
        cy.contains('a', 'Return to task list').click()

        cy.contains('a', 'Check and send to your certificate provider').click()
        cy.url().should('contain', '/you-cannot-sign-your-lpa-yet')

        cy.scrollTo('bottom');
        cy.contains('.govuk-summary-list__row', 'Jessie Jones').contains('a', 'Change').click();

        cy.url().should('contain', '/enter-attorney')
        cy.get('#f-date-of-birth-year').invoke('val', "2000");
        cy.contains('button', 'Save and continue').click()
        cy.url().should('contain', '/you-cannot-sign-your-lpa-yet')

        cy.contains('.govuk-summary-list__row', 'Blake Buckley').contains('a', 'Change').click();

        cy.url().should('contain', '/enter-replacement-attorney')
        cy.get('#f-date-of-birth-year').invoke('val', "2000");
        cy.contains('button', 'Save and continue').click()
        cy.url().should('contain', '/task-list')
    });
});
