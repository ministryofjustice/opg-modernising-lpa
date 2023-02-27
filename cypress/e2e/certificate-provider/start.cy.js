describe('Start', () => {
    beforeEach(() => {
        cy.visit('/testing-start?startCpFlowDonorHasPaid=1');
    });

    it('can be completed', () => {
        cy.contains('a', 'Start');
    });
});
