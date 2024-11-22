// TODO unskip MLPAB-2700
describe.skip('You cannot sign your LPA yet', () => {
    it('lists attorneys and replacement attorneys that are under 18', () => {
        const today = new Date()
        cy.visit('/fixtures?redirect=/choose-attorneys-summary&progress=peopleToNotifyAboutYourLpa');

        cy.contains('.govuk-summary-card', 'Jessie Jones').contains('a', 'Change').click();
        cy.get('#f-date-of-birth-year').clear().type(today.getFullYear() - 1)
        cy.contains('button', 'Save and continue').click()
        cy.contains('button', 'Save and continue').click()
        cy.visitLpa('/choose-replacement-attorneys-summary')

        cy.contains('.govuk-summary-card', 'Blake Buckley').contains('a', 'Change').click();
        cy.get('#f-date-of-birth-year').clear().type(today.getFullYear() - 1)
        cy.contains('button', 'Save and continue').click()
        cy.contains('button', 'Save and continue').click()
        cy.contains('a', 'Return to task list').click()

        cy.contains('a', 'Check and send to your certificate provider').click()
        cy.url().should('contain', '/you-cannot-sign-your-lpa-yet')

        cy.contains('.govuk-summary-list__row', 'Jessie Jones').contains('a', 'Change').click();

        cy.url().should('contain', '/choose-attorneys')
        cy.get('#f-date-of-birth-year').clear().type("2000")
        cy.contains('button', 'Save and continue').click()
        cy.url().should('contain', '/you-cannot-sign-your-lpa-yet')

        cy.contains('.govuk-summary-list__row', 'Blake Buckley').contains('a', 'Change').click();

        cy.url().should('contain', '/choose-replacement-attorneys')
        cy.get('#f-date-of-birth-year').clear().type("2000")
        cy.contains('button', 'Save and continue').click()
        cy.url().should('contain', '/task-list')
    });
});
