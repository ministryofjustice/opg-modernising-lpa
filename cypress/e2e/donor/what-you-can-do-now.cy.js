describe('what you can do now', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/what-is-vouching&progress=confirmYourIdentity&IdStatus=insufficient-evidence')
        cy.url().should('contain', '/what-is-vouching')
        cy.checkA11yApp()

        cy.get('input[name="yes-no"]').check('no', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/what-you-can-do-now')
    })

    it('can choose to get ID documents', () => {
        cy.get('input[name="do-next"]').check('prove-own-id', { force: true });
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/task-list')
    })

    it('can choose to add a voucher', () => {
        cy.get('input[name="do-next"]').check('select-new-voucher', { force: true });
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/enter-voucher')
    })

    it('can choose to withdraw LPA', () => {
        cy.get('input[name="do-next"]').check('withdraw-lpa', { force: true });
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/withdraw-this-lpa')
    })

    it('can choose to apply to court of protection', () => {
        cy.get('input[name="do-next"]').check('apply-to-cop', { force: true });
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/register-with-court-of-protection')
    })

    it('errors when option not selected', () => {
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/what-you-can-do-now')
        cy.checkA11yApp()

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select what you would like to do');
        });

        cy.contains('.govuk-error-message', 'Select what you would like to do');
    })
})
