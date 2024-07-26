describe('View LPA', () => {
    beforeEach(() => {
        cy.visit('/fixtures/supporter?organisation=1&redirect=/dashboard&lpa=1');
        cy.checkA11yApp();
    });

    it('can continue making an LPA', () => {
        cy.contains('a', 'M-FAKE').click()

        cy.url().should('contain', '/view-lpa');
        cy.checkA11yApp();

        cy.contains('h1', 'Property and affairs LPA')
        cy.contains('div', 'M-FAKE')

        cy.contains('a', 'Donor access')
        cy.contains('a', 'View LPA summary')
        cy.contains('a', 'Go to task list').click()

        cy.url().should('contain', '/task-list');
        cy.checkA11yApp();

        cy.contains('Provide your details').click()

        cy.url().should('contain', '/your-details');
        cy.checkA11yApp();

        cy.contains('.govuk-summary-list__row', 'First names').find('a').click();

        cy.get('#f-first-names').clear().type('2');
        cy.contains('button', 'Save and continue').click();
        cy.contains('a', 'Dashboard').click();
        cy.contains('2 Smith');
    });

    it('shows progress of LPA', () => {
        cy.contains('a', 'M-FAKE').click()

        cy.contains('h2', 'LPA progress')

        cy.contains('li', 'Sam Smith has paid In progress')
        cy.contains('li', 'Sam Smith has confirmed their identity Not started')
        cy.contains('li', 'Sam Smith has signed the LPA Not started')
        cy.contains('li', 'The certificate provider has provided their certificate Not started')
        cy.contains('li', 'All attorneys have signed the LPA Not started')
        cy.contains('li', 'OPG has received the LPA Not started')
        cy.contains('li', 'The 4-week waiting period has started Not started')
        cy.contains('li', 'The LPA has been registered Not started')

        cy.visit('/fixtures/supporter?organisation=1&redirect=/dashboard&lpa=1&setLPAProgress=1&progress=registered');

        cy.contains('a', 'M-FAKE').click()

        cy.contains('li', 'Sam Smith has paid Completed')
        cy.contains('li', 'Sam Smith has confirmed their identity Completed')
        cy.contains('li', 'Sam Smith has signed the LPA Completed')
        cy.contains('li', 'The certificate provider has provided their certificate')
        cy.contains('li', 'All attorneys have signed the LPA Completed')
        cy.contains('li', 'OPG has received the LPA Completed')
        cy.contains('li', 'The 4-week waiting period has started Completed')
        cy.contains('li', 'The LPA has been registered Completed')
    })
})
