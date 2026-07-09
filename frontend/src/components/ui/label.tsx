import * as React from 'react';
import { cn } from '@/lib/utils';

export interface LabelProps extends React.LabelHTMLAttributes<HTMLLabelElement> {}

const Label = React.forwardRef<HTMLLabelElement, LabelProps>(({ className, ...props }, ref) => {
  return (
    <label
      className={cn(
        'flex items-center gap-1.5 text-sm font-medium leading-none text-foreground select-none',
        className,
      )}
      ref={ref}
      {...props}
    />
  );
});
Label.displayName = 'Label';

export { Label };
