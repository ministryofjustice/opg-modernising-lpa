describe('Start', () => {
    beforeEach(() => {
        cy.visit('/testing-start?startCpFlowWithoutId=1');
    });

    it('can be completed', () => {
        cy.contains('a', 'Start');
    });
});
