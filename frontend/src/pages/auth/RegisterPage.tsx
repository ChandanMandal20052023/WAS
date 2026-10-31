import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Lock, Mail, User, Building2, ArrowRight, AlertCircle, Loader2 } from 'lucide-react';

import { authService } from '../../services/authService';
import { useAuthStore } from '../../stores/authStore';

const registerSchema = z.object({
  business_name: z.string().min(2, 'Business name must be at least 2 characters'),
  full_name: z.string().min(2, 'Full name must be at least 2 characters'),
  email: z.string().email('Please enter a valid work email address'),
  password: z.string().min(8, 'Password must be at least 8 characters long'),
  confirm_password: z.string(),
}).refine((data) => data.password === data.confirm_password, {
  message: 'Passwords do not match',
  path: ['confirm_password'],
});

type RegisterFormValues = z.infer<typeof registerSchema>;

export const RegisterPage: React.FC = () => {
  const navigate = useNavigate();
  const setAuth = useAuthStore((state) => state.setAuth);

  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (values: RegisterFormValues) => {
    try {
      setIsSubmitting(true);
      setErrorMessage(null);
      const response = await authService.register({
        business_name: values.business_name,
        full_name: values.full_name,
        email: values.email,
        password: values.password,
      });
      setAuth(response);
      navigate('/', { replace: true });
    } catch (err: unknown) {
      if (err instanceof Error) {
        setErrorMessage(err.message);
      } else {
        setErrorMessage('Registration failed. Please try again.');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen w-full flex items-center justify-center bg-background px-4 py-12">
      <div className="w-full max-w-md space-y-6 bg-card border border-border p-8 rounded-2xl shadow-lg">
        {/* Branding Header */}
        <div className="text-center space-y-2">
          <div className="inline-flex w-12 h-12 rounded-xl bg-whatsapp items-center justify-center text-white font-bold text-2xl shadow-md">
            W
          </div>
          <h2 className="text-2xl font-bold tracking-tight text-foreground">Create your organization</h2>
          <p className="text-sm text-muted-foreground">
            Get started with WAS to centralize your customer WhatsApp conversations.
          </p>
        </div>

        {/* Error Alert */}
        {errorMessage && (
          <div className="flex items-center gap-2 p-3 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-xl">
            <AlertCircle className="w-4 h-4 flex-shrink-0" />
            <span>{errorMessage}</span>
          </div>
        )}

        {/* Registration Form */}
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-3.5">
          <div className="space-y-1">
            <label className="text-xs font-semibold text-foreground">Business / Organization Name</label>
            <div className="relative">
              <Building2 className="w-4 h-4 absolute left-3 top-3 text-muted-foreground" />
              <input
                {...register('business_name')}
                type="text"
                placeholder="Acme Global Ltd"
                className="w-full bg-background border border-input pl-9 pr-3 py-2 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
              />
            </div>
            {errors.business_name && (
              <p className="text-xs text-destructive">{errors.business_name.message}</p>
            )}
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-foreground">Your Full Name</label>
            <div className="relative">
              <User className="w-4 h-4 absolute left-3 top-3 text-muted-foreground" />
              <input
                {...register('full_name')}
                type="text"
                placeholder="John Doe"
                className="w-full bg-background border border-input pl-9 pr-3 py-2 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
              />
            </div>
            {errors.full_name && (
              <p className="text-xs text-destructive">{errors.full_name.message}</p>
            )}
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-foreground">Work Email</label>
            <div className="relative">
              <Mail className="w-4 h-4 absolute left-3 top-3 text-muted-foreground" />
              <input
                {...register('email')}
                type="email"
                placeholder="john@acme.com"
                className="w-full bg-background border border-input pl-9 pr-3 py-2 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
              />
            </div>
            {errors.email && (
              <p className="text-xs text-destructive">{errors.email.message}</p>
            )}
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-foreground">Password (min. 8 characters)</label>
            <div className="relative">
              <Lock className="w-4 h-4 absolute left-3 top-3 text-muted-foreground" />
              <input
                {...register('password')}
                type="password"
                placeholder="••••••••"
                className="w-full bg-background border border-input pl-9 pr-3 py-2 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
              />
            </div>
            {errors.password && (
              <p className="text-xs text-destructive">{errors.password.message}</p>
            )}
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-foreground">Confirm Password</label>
            <div className="relative">
              <Lock className="w-4 h-4 absolute left-3 top-3 text-muted-foreground" />
              <input
                {...register('confirm_password')}
                type="password"
                placeholder="••••••••"
                className="w-full bg-background border border-input pl-9 pr-3 py-2 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
              />
            </div>
            {errors.confirm_password && (
              <p className="text-xs text-destructive">{errors.confirm_password.message}</p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full mt-2 bg-whatsapp hover:bg-whatsapp-dark text-white font-medium py-2.5 px-4 rounded-xl text-sm flex items-center justify-center gap-2 transition-all shadow-sm hover:shadow disabled:opacity-50"
          >
            {isSubmitting ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <>
                <span>Create Organization & Start</span>
                <ArrowRight className="w-4 h-4" />
              </>
            )}
          </button>
        </form>

        {/* Footer Link */}
        <div className="text-center pt-2 border-t border-border">
          <p className="text-xs text-muted-foreground">
            Already have an account?{' '}
            <Link to="/login" className="text-primary font-semibold hover:underline">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
};
