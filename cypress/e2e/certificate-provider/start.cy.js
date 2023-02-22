describe('Start', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/your-details&withIncompleteAttorneys=1&withCP=1&paymentComplete=1');
        cy.visitLpa('/task-list');
    });

    it('can be completed', () => {
        cy.contains('a', 'Start');
    });
});
