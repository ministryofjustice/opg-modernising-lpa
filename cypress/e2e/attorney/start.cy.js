describe('Start', () => {
    beforeEach(() => {
        cy.visit('/attorney-start');
    });

    it('can be started', () => {
        cy.contains('a', 'Start');
    });
});
