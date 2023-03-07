describe('Restrictions', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/restrictions&withDonorDetails=1&withAttorney=1');
    });

    it('can be submitted', () => {
        cy.get('#f-restrictions').type('this that');

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/who-do-you-want-to-be-certificate-provider-guidance');
    });
});
