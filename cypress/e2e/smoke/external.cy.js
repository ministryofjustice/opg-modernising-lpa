describe('external dependencies', () => {
    describe('UID service', () => {
        it('request signing and base URL are configured correctly', () => {
            cy.request('/health-check/dependency').should((response) => {
                expect(response.status).to.eq(200)
            })
        })
    })
})
