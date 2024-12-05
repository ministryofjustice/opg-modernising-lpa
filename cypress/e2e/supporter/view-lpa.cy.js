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

        cy.contains('li', 'LPA paid for Not completed')
        cy.contains('li', 'Sam Smith’s identity confirmed Not completed')
        cy.contains('li', 'LPA signed by Sam Smith Not completed')
        cy.contains('li', 'LPA certificate provided by Charlie Cooper Not completed')
        cy.contains('li', 'LPA signed by all attorneys Not completed')
        cy.contains('li', 'OPG’s statutory 4-week waiting period begins Not completed')
        cy.contains('li', 'Sam Smith’s LPA registered by OPG Not completed')

        cy.visit('/fixtures/supporter?organisation=1&redirect=/dashboard&lpa=1&setLPAProgress=1&progress=registered');

        cy.contains('a', 'M-FAKE').click()

        cy.contains('li', 'LPA paid for Completed')
        cy.contains('li', 'Sam Smith’s identity confirmed Completed')
        cy.contains('li', 'LPA signed by Sam Smith Completed')
        cy.contains('li', 'LPA certificate provided by Charlie Cooper Completed')
        cy.contains('li', 'LPA signed by all attorneys Completed')
        cy.contains('li', 'OPG’s statutory 4-week waiting period begins Completed')
        cy.contains('li', 'Sam Smith’s LPA registered by OPG Completed')
    })
})
