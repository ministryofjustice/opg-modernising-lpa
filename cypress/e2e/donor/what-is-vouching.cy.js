describe('what is vouching', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/what-is-vouching&progress=payForTheLpa')
        cy.url().should('contain', '/what-is-vouching')
        cy.checkA11yApp()
    })

    it('can confirm has a voucher', () => {
        cy.get('input[name="yes-no"]').check('yes', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/enter-voucher')
    })

    it('can confirm has not got a voucher', () => {
        cy.get('input[name="yes-no"]').check('no', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/task-list')
    })

    it('errors when option not selected', () => {
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/what-is-vouching')
        cy.checkA11yApp()

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select yes if you know someone and they have agreed to vouch for you');
        });

        cy.contains('.govuk-error-message', 'Select yes if you know someone and they have agreed to vouch for you');
    })
})
