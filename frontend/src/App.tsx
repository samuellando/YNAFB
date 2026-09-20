import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Route, Routes } from 'react-router'
import Home from './pages/Home'
import Login from './pages/Login'
import Signup from './pages/Signup'
import Budget from './pages/Budget'
import MonthlyBudget from './pages/MonthlyBudget'
import Account from './pages/Account'
import ExpenseShare from './pages/ExpenseShare'
import AppLayout from './components/AppLayout'
import RequireSession from './components/RequireSession'

const queryClient = new QueryClient()

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />} />
          <Route path="/signup" element={<Signup />} />
          <Route path="/budget" element={<RequireSession />}>
            <Route index element={<Budget />} />
            <Route path=":budgetId" element={<AppLayout />}>
              <Route index element={<MonthlyBudget />} />
              <Route path=":month" element={<MonthlyBudget />} />
              <Route path="account/:accountId" element={<Account />} />
              <Route path="expense-share/:expenseShareId" element={<ExpenseShare />} />
            </Route>
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

export default App
